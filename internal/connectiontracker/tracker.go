package connectiontracker

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const snapLen = 80 // enough to read ipv6 but parsing only ipv4

// quick reference https://serverfault.com/a/1000310
// capturing inbound and outbound traffic
const bpfFilter = `tcp[tcpflags] &(tcp-syn) != 0 and tcp[tcpflags] &(tcp-ack) = 0`

var (
	counter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tcptracker_new_connections",
		Help: "The current number of tracked requests, take a look at rate(tcptracker_new_connections[5m])",
	})
)

// ConnEntry is used for tracking the connections SourceIP -> DestinationIP : map/set of destination Ports[]
// `Add the ability to detect a port scan, where a single source IP connects to more than 3 host Ports in the previous minute.`
type ConnEntry struct {
	SrcIP *net.IP
	DstIP *net.IP
	Ports map[int]bool
}

var (
	ethLayer layers.Ethernet
	ipLayer  layers.IPv4
	tcpLayer layers.TCP
	tlsLayer layers.TLS
	udpLayer layers.UDP
)

// Tracker contains methods to track Connections and Block IPs
type Tracker struct {
	deviceName       string
	cache            *connCache
	bpfFilter        string
	parser           *gopacket.DecodingLayerParser
	snapLen          int
	minimumPortScans int
	firewall         Firewall
	timeout          time.Duration
}

// TrackerParams required params to run Tracker
type TrackerParams struct {
	DeviceName string
	Firewall   Firewall
	Metrics    *prometheus.Registry
}

func NewTracker(p TrackerParams) *Tracker {
	p.Metrics.MustRegister(counter)
	// TODO: pass values via config, env vars
	return &Tracker{
		deviceName:       p.DeviceName,
		cache:            newCacheManager(1 * time.Minute),
		bpfFilter:        bpfFilter,
		snapLen:          snapLen,
		parser:           newPacketParser(),
		minimumPortScans: 3,
		firewall:         p.Firewall,
		timeout:          pcap.BlockForever,
	}
}

func newPacketParser() *gopacket.DecodingLayerParser {
	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&ethLayer,
		&ipLayer,
		&tcpLayer,
		&tlsLayer,
		&udpLayer,
	)
	return parser
}

func (t *Tracker) Execute(ctx context.Context) {
	newConnections := make(chan *ConnEntry)
	defer close(newConnections)
	portScans := make(chan *ConnEntry)
	defer close(portScans)

	go t.trackConnections(ctx, newConnections, portScans)
	go t.onDetectedPortScan(portScans)
	t.capture(newConnections)
}

// capture is using gopacket lib to capture connections and send it to another channel
func (t *Tracker) capture(newConnections chan *ConnEntry) {
	log.Info().Msg("TCPTracker: capture is running...")
	handle, err := pcap.OpenLive(t.deviceName, snapLen, false, t.timeout)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer handle.Close()
	errBPF := handle.SetBPFFilter(t.bpfFilter)
	if errBPF != nil {
		log.Fatal().Err(errBPF).Send()
	}
	var foundLayerTypes []gopacket.LayerType
	source := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range source.Packets() {
		err := t.parser.DecodeLayers(packet.Data(), &foundLayerTypes)
		if err != nil {
			log.Error().Err(err)
			return
		}
		ip4, tcp := decodeLayers(packet)
		counter.Inc()
		newConnections <- prepareEntry(ip4, tcp)
	}
}

func prepareEntry(ip4 *layers.IPv4, tcp *layers.TCP) *ConnEntry {
	log.Info().Msgf("New connection: %s:%v -> %s:%v", ip4.SrcIP, int(tcp.SrcPort), ip4.DstIP, int(tcp.DstPort))
	entry := &ConnEntry{
		SrcIP: &ip4.SrcIP,
		DstIP: &ip4.DstIP,
		Ports: map[int]bool{
			int(tcp.DstPort): true,
		},
	}
	return entry

}

func decodeLayers(packet gopacket.Packet) (*layers.IPv4, *layers.TCP) {
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ipv4Layer == nil {
		log.Error().Msgf("IPv4Layer is nil")
	}
	tcpTypeLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpTypeLayer == nil {
		log.Error().Msgf("TCPTypeLayer is nil")
	}
	ip4 := ipv4Layer.(*layers.IPv4)
	tcp := tcpTypeLayer.(*layers.TCP)
	return ip4, tcp
}

// trackConnections is getting new connections from capture and checking isPortScanning
func (t *Tracker) trackConnections(ctx context.Context, newConnections chan *ConnEntry, portScans chan *ConnEntry) {
	log.Info().Msg("TCPTracker: trackConnections is running...")
	var wg sync.WaitGroup
	for conn := range newConnections {
		log.Info().Msgf("Tracking connection from %s:%s", conn.SrcIP.String(), intMapToString(conn.Ports))
		wg.Add(1)
		go func(conn *ConnEntry) {
			wg.Done()
			found := t.cache.getOrSet(ctx, conn)
			if isPortScanning(len(found.Ports), t.minimumPortScans) {
				portScans <- found
			}
		}(conn)
	}
	wg.Wait()

}

func isPortScanning(foundPorts, minimumPortScans int) bool {
	return foundPorts > minimumPortScans
}

// onDetectedPortScan is blocking IP in Host Firewall
func (t *Tracker) onDetectedPortScan(portScans chan *ConnEntry) {
	log.Info().Msg("TCPTracker: onDetectedPortScan is running...")
	for v := range portScans {
		if v.SrcIP == nil {
			return
		}
		log.Info().Msgf("TCPTracker: Port scan detected: %s -> %s on Ports %v", v.SrcIP, v.DstIP, intMapToString(v.Ports))
		ip := v.SrcIP.String()
		err := t.firewall.Block(ip)
		if err != nil {
			log.Err(err).Send()
		}
		log.Info().Msgf("TCPTracker: Blocked %s IP", v.SrcIP)
	}
}

func intMapToString(portsMap map[int]bool) string {
	ports := make([]string, 0, len(portsMap))
	for k := range portsMap {
		ports = append(ports, strconv.Itoa(k))
	}
	sort.Strings(ports)
	return strings.Join(ports, ",")
}

func (t *Tracker) Close() error {
	err := t.firewall.Close()
	if err != nil {
		return err
	}
	return nil
}
