package connectiontracker

import (
	"github.com/coreos/go-iptables/iptables"
	"github.com/google/gopacket/pcap"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
	"net"
)

const (
	inputChain = "INPUT"
	// separate chain for isolation
	trackerChain = "tcptracker"
	table        = "filter"
	drop         = "DROP"
)

//Firewall interface for Blocking IPs
//go:generate mockgen -source=firewall.go -package=mock -destination=../../mock/gomock_firewall.go Firewall
type Firewall interface {
	Block(ip string) error
	Close() error
}

//go:generate mockgen -source=firewall.go -package=mock -destination=../../mock/gomock_ipTableCoreos.go ipTableCoreos
// ipTableCoreos is matching implementation of coreos/go-iptables/iptables
type ipTableCoreos interface {
	ChainExists(string, string) (bool, error)
	NewChain(string, string) error
	Insert(string, string, int, ...string) error
	AppendUnique(string, string, ...string) error
	DeleteIfExists(string, string, ...string) error
	ClearAndDeleteChain(string, string) error
}

// IPTables gives functionalities to Block IP addresses
type IPTables struct {
	iptables     ipTableCoreos
	jumpRuleSpec []string
	allowList    []string // TODO: something to investigate more
}

// NewFirewall returns and instance of IPTables
func NewFirewall(deviceName string) (Firewall, error) {
	localIP, ok := getLocalIP(deviceName)
	if !ok {
		log.Fatal().Msgf("Cannot track packets for non existing device: %s", deviceName)
	}
	ipv4, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return nil, err
	}
	fw := newFW(ipv4, getLocalIPString(localIP))
	errInit := initialise(fw)
	if errInit != nil {
		return nil, errInit
	}
	return fw, nil
}

func getLocalIPString(localIP net.IP) string {
	localIPString := ""
	if localIP != nil {
		localIPString = localIP.String()
	}
	return localIPString
}

func newFW(ipv4 *iptables.IPTables, localIP string) *IPTables {
	fw := &IPTables{
		iptables:     ipv4,
		jumpRuleSpec: []string{"-m", "state", "--state", "NEW", "-j", trackerChain},
		allowList:    []string{localIP},
	}
	return fw
}

// Block takes the IP address and adding it to chain as DROP = block
func (fw *IPTables) Block(ip string) error {
	rule := []string{"-s", ip, "-j", drop}
	if slices.Contains(fw.allowList, ip) {
		log.Warn().Msgf("%s IP is on the allow list... skipping...", ip)
		return nil
	}
	// AppendUnique acts like Append except that it won't add a duplicate
	return fw.iptables.AppendUnique(table, trackerChain, rule...)
}

func initialise(fw *IPTables) error {
	if err := clear(fw.iptables, fw.jumpRuleSpec); err != nil {
		return err
	}
	if errCreate := create(fw.iptables, fw.jumpRuleSpec); errCreate != nil {
		return errCreate
	}
	return nil
}

func create(iptables ipTableCoreos, jumpRuleSpec []string) error {
	ok, err := iptables.ChainExists(table, trackerChain)
	if err != nil {
		return err
	}
	if !ok {
		if err := iptables.NewChain(table, trackerChain); err != nil {
			return err
		}
	}
	if err := iptables.Insert(table, inputChain, 1, jumpRuleSpec...); err != nil {
		return err
	}
	return nil

}

func clear(iptables ipTableCoreos, jumpRuleSpec []string) error {
	ok, err := iptables.ChainExists(table, trackerChain)
	if err != nil {
		return err
	}
	if ok {
		if err := iptables.DeleteIfExists(table, inputChain, jumpRuleSpec...); err != nil {
			return err
		}
		if err := iptables.ClearAndDeleteChain(table, trackerChain); err != nil {
			return err
		}
	}
	return nil
}

func getLocalIP(deviceName string) (net.IP, bool) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	for _, d := range devices {
		if d.Name == deviceName {
			for _, addr := range d.Addresses {
				return addr.IP.To4(), true
			}
		}
	}
	return nil, false
}

func (fw *IPTables) Close() error {
	return clear(fw.iptables, fw.jumpRuleSpec)
}
