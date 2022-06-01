package connectiontracker

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"net"
	"sync"
	"tcptracker/mock"
	"testing"
	"time"
)

var testSimpleTCPPacket = []byte{
	0x00, 0x00, 0x0c, 0x9f, 0xf0, 0x20, 0xbc, 0x30, 0x5b, 0xe8, 0xd3, 0x49,
	0x08, 0x00, 0x45, 0x00, 0x01, 0xa4, 0x39, 0xdf, 0x40, 0x00, 0x40, 0x06,
	0x55, 0x5a, 0xac, 0x11, 0x51, 0x49, 0xad, 0xde, 0xfe, 0xe1, 0xc5, 0xf7,
	0x00, 0x50, 0xc5, 0x7e, 0x0e, 0x48, 0x49, 0x07, 0x42, 0x32, 0x80, 0x18,
	0x00, 0x73, 0x9a, 0x8f, 0x00, 0x00, 0x01, 0x01, 0x08, 0x0a, 0x03, 0x77,
	0x37, 0x9c, 0x42, 0x77, 0x5e, 0x3a, 0x47, 0x45, 0x54, 0x20, 0x2f, 0x20,
	0x48, 0x54, 0x54, 0x50, 0x2f, 0x31, 0x2e, 0x31, 0x0d, 0x0a, 0x48, 0x6f,
	0x73, 0x74, 0x3a, 0x20, 0x77, 0x77, 0x77, 0x2e, 0x66, 0x69, 0x73, 0x68,
	0x2e, 0x63, 0x6f, 0x6d, 0x0d, 0x0a, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x3a, 0x20, 0x6b, 0x65, 0x65, 0x70, 0x2d, 0x61,
	0x6c, 0x69, 0x76, 0x65, 0x0d, 0x0a, 0x55, 0x73, 0x65, 0x72, 0x2d, 0x41,
	0x67, 0x65, 0x6e, 0x74, 0x3a, 0x20, 0x4d, 0x6f, 0x7a, 0x69, 0x6c, 0x6c,
	0x61, 0x2f, 0x35, 0x2e, 0x30, 0x20, 0x28, 0x58, 0x31, 0x31, 0x3b, 0x20,
	0x4c, 0x69, 0x6e, 0x75, 0x78, 0x20, 0x78, 0x38, 0x36, 0x5f, 0x36, 0x34,
	0x29, 0x20, 0x41, 0x70, 0x70, 0x6c, 0x65, 0x57, 0x65, 0x62, 0x4b, 0x69,
	0x74, 0x2f, 0x35, 0x33, 0x35, 0x2e, 0x32, 0x20, 0x28, 0x4b, 0x48, 0x54,
	0x4d, 0x4c, 0x2c, 0x20, 0x6c, 0x69, 0x6b, 0x65, 0x20, 0x47, 0x65, 0x63,
	0x6b, 0x6f, 0x29, 0x20, 0x43, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2f, 0x31,
	0x35, 0x2e, 0x30, 0x2e, 0x38, 0x37, 0x34, 0x2e, 0x31, 0x32, 0x31, 0x20,
	0x53, 0x61, 0x66, 0x61, 0x72, 0x69, 0x2f, 0x35, 0x33, 0x35, 0x2e, 0x32,
	0x0d, 0x0a, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x3a, 0x20, 0x74, 0x65,
	0x78, 0x74, 0x2f, 0x68, 0x74, 0x6d, 0x6c, 0x2c, 0x61, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x78, 0x68, 0x74, 0x6d,
	0x6c, 0x2b, 0x78, 0x6d, 0x6c, 0x2c, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x78, 0x6d, 0x6c, 0x3b, 0x71, 0x3d,
	0x30, 0x2e, 0x39, 0x2c, 0x2a, 0x2f, 0x2a, 0x3b, 0x71, 0x3d, 0x30, 0x2e,
	0x38, 0x0d, 0x0a, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x2d, 0x45, 0x6e,
	0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x3a, 0x20, 0x67, 0x7a, 0x69, 0x70,
	0x2c, 0x64, 0x65, 0x66, 0x6c, 0x61, 0x74, 0x65, 0x2c, 0x73, 0x64, 0x63,
	0x68, 0x0d, 0x0a, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x2d, 0x4c, 0x61,
	0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x3a, 0x20, 0x65, 0x6e, 0x2d, 0x55,
	0x53, 0x2c, 0x65, 0x6e, 0x3b, 0x71, 0x3d, 0x30, 0x2e, 0x38, 0x0d, 0x0a,
	0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x2d, 0x43, 0x68, 0x61, 0x72, 0x73,
	0x65, 0x74, 0x3a, 0x20, 0x49, 0x53, 0x4f, 0x2d, 0x38, 0x38, 0x35, 0x39,
	0x2d, 0x31, 0x2c, 0x75, 0x74, 0x66, 0x2d, 0x38, 0x3b, 0x71, 0x3d, 0x30,
	0x2e, 0x37, 0x2c, 0x2a, 0x3b, 0x71, 0x3d, 0x30, 0x2e, 0x33, 0x0d, 0x0a,
	0x0d, 0x0a,
}

func Test_intMapToString(t *testing.T) {
	type args struct {
		portsMap map[int]bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "successfully convert",
			args: args{
				portsMap: map[int]bool{
					7070: true,
					8080: true,
					9090: true,
				},
			},
			want: "7070,8080,9090",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, intMapToString(tt.args.portsMap), "intMapToString(%v)", tt.args.portsMap)
		})
	}
}

func Test_decodeLayers(t *testing.T) {
	// example from gopacket lib
	type args struct {
		packet gopacket.Packet
	}
	tests := []struct {
		name       string
		args       args
		ip4Src     string
		ip4Dst     string
		tcpSrcPort int
		tcpDstPort int
	}{
		{
			name: "decode success",
			args: args{
				packet: gopacket.NewPacket(testSimpleTCPPacket, layers.LinkTypeEthernet, gopacket.DecodeOptions{Lazy: true, NoCopy: true}),
			},
			ip4Src:     "172.17.81.73",
			ip4Dst:     "173.222.254.225",
			tcpSrcPort: 50679,
			tcpDstPort: 80,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipv4, tcp := decodeLayers(tt.args.packet)
			assert.Equalf(t, tt.tcpSrcPort, int(tcp.SrcPort), "decodeLayers(%v)", tt.args.packet)
			assert.Equalf(t, tt.tcpDstPort, int(tcp.DstPort), "decodeLayers(%v)", tt.args.packet)
			assert.Equalf(t, tt.ip4Src, ipv4.SrcIP.String(), "decodeLayers(%v)", tt.args.packet)
			assert.Equalf(t, tt.ip4Dst, ipv4.DstIP.String(), "decodeLayers(%v)", tt.args.packet)

		})
	}
}

func Test_Tracker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFw := mock.NewMockFirewall(ctrl)
	params := TrackerParams{
		DeviceName: "eth0",
		Firewall:   mockFw,
		Metrics:    prometheus.NewRegistry(),
	}
	tracker := NewTracker(params)
	require.NotNil(t, tracker)

	dstIP := net.ParseIP("192.44.55.66")
	srcIP := net.ParseIP("172.44.55.76")
	ip := srcIP.String()
	mockFw.EXPECT().Block(gomock.Eq(ip)).Return(nil).Times(1)

	newConnections := make(chan *ConnEntry, 4)
	testPortScans := make(chan *ConnEntry, 1)
	portScans := make(chan *ConnEntry, 1)
	done := make(chan int)
	donePortChan := make(chan int)

	var wg sync.WaitGroup
	conns := generateConns(srcIP, dstIP)
	go tracker.trackConnections(context.Background(), newConnections, testPortScans)
	wg.Add(1)
	go produce(t, newConnections, done, &wg, conns)
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	close(newConnections)
	close(testPortScans)

	portScanDetected := connWithScanDetected(srcIP, dstIP)
	go tracker.onDetectedPortScan(portScans)

	wg.Add(1)
	go produce(t, portScans, donePortChan, &wg, portScanDetected)
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	close(portScans)

	// TODO structure code better to make it easier to test or find a better pattern and synchronize it better in tests

}

func connWithScanDetected(srcIP net.IP, dstIP net.IP) []*ConnEntry {
	portScanDetected := []*ConnEntry{
		{
			SrcIP: &srcIP,
			DstIP: &dstIP,
			Ports: map[int]bool{
				7070: true,
				8080: true,
				9090: true,
				9191: true,
			},
		},
	}
	return portScanDetected
}

func produce(t *testing.T, ch chan *ConnEntry, quit chan int, wg *sync.WaitGroup, conns []*ConnEntry) {
	defer wg.Done()
	for _, c := range conns {
		select {
		case <-quit:
			close(ch)
			fmt.Println("exit")

			return
		case ch <- c:
			assert.Equal(t, "172.44.55.76", c.SrcIP.String())
			fmt.Println("Producer sends", c)
		}
	}
}

func generateConns(srcIP net.IP, dstIP net.IP) []*ConnEntry {
	conn1 := &ConnEntry{
		SrcIP: &srcIP,
		DstIP: &dstIP,
		Ports: map[int]bool{
			7070: true,
		},
	}
	conn2 := &ConnEntry{
		SrcIP: &srcIP,
		DstIP: &dstIP,
		Ports: map[int]bool{
			8080: true,
		},
	}
	conn3 := &ConnEntry{
		SrcIP: &srcIP,
		DstIP: &dstIP,
		Ports: map[int]bool{
			9090: true,
		},
	}
	conn4 := &ConnEntry{
		SrcIP: &srcIP,
		DstIP: &dstIP,
		Ports: map[int]bool{
			9191: true,
		},
	}
	return []*ConnEntry{conn1, conn2, conn3, conn4}
}

func Test_prepareEntry(t *testing.T) {
	packet := gopacket.NewPacket(testSimpleTCPPacket, layers.LinkTypeEthernet, gopacket.DecodeOptions{Lazy: true, NoCopy: true})
	ipv4, tcp := decodeLayers(packet)
	entry := prepareEntry(ipv4, tcp)
	assert.NotNil(t, entry)
	assert.Equal(t, "172.17.81.73", entry.SrcIP.String())
	assert.Equal(t, "173.222.254.225", entry.DstIP.String())
	assert.Equal(t, 80, maps.Keys(entry.Ports)[0])
	assert.True(t, entry.Ports[80])
}
