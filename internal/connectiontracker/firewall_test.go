package connectiontracker

import (
	"github.com/coreos/go-iptables/iptables"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	mock2 "tcptracker/mock"
	"testing"
)

func Test_deviceExists(t *testing.T) {
	_, notExiting := getLocalIP("notExiting")
	assert.False(t, notExiting)

	// check any of these popular interfaces
	ip1, existsEth0 := getLocalIP("eth0")
	ip2, existsEno1 := getLocalIP("eno1")
	assert.True(t, existsEno1 || existsEth0)
	if ip1 != nil {
		assert.NotEmpty(t, ip1.String())
	}
	if ip2 != nil {
		assert.NotEmpty(t, ip2.String())
	}
}

func TestBlockIpv4(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIptables := mock2.NewMockIptablesMock(mockCtrl)

	ipAllowed := "192.169.0.2"
	firewall := IPTables{
		iptables:     mockIptables,
		jumpRuleSpec: []string{"-m", "state", "--state", "NEW", "-j", trackerChain},
		allowList:    []string{ipAllowed},
	}
	ip := "192.169.0.1"
	mockIptables.EXPECT().AppendUnique(table, trackerChain, []string{"-s", ip, "-j", drop}).Return(nil).Times(1)
	err := firewall.Block(ip)
	errAllowed := firewall.Block(ipAllowed)
	require.NoError(t, err)
	require.NoError(t, errAllowed)
}

func TestClearWhenExists(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIptables := mock2.NewMockIptablesMock(mockCtrl)

	firewall := IPTables{
		iptables:     mockIptables,
		jumpRuleSpec: []string{"-m", "state", "--state", "NEW", "-j", trackerChain},
	}
	mockIptables.EXPECT().ChainExists(table, trackerChain).Return(true, nil).Times(1)
	mockIptables.EXPECT().DeleteIfExists(table, inputChain, firewall.jumpRuleSpec).Return(nil).Times(1)
	mockIptables.EXPECT().ClearAndDeleteChain(table, trackerChain).Return(nil).Times(1)

	err := clear(mockIptables, firewall.jumpRuleSpec)
	require.NoError(t, err)
}

func TestClearOnNonExistingChain(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIptables := mock2.NewMockIptablesMock(mockCtrl)

	firewall := IPTables{
		iptables:     mockIptables,
		jumpRuleSpec: []string{"-m", "state", "--state", "NEW", "-j", trackerChain},
	}
	mockIptables.EXPECT().ChainExists(table, trackerChain).Return(false, nil).Times(1)
	mockIptables.EXPECT().DeleteIfExists(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	mockIptables.EXPECT().ClearAndDeleteChain(gomock.Any(), gomock.Any()).Times(0)

	err := clear(mockIptables, firewall.jumpRuleSpec)
	require.NoError(t, err)
}

func TestCreateInteractions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIptables := mock2.NewMockIptablesMock(mockCtrl)

	firewall := IPTables{
		iptables:     mockIptables,
		jumpRuleSpec: []string{"-m", "state", "--state", "NEW", "-j", trackerChain},
	}
	mockIptables.EXPECT().ChainExists(table, trackerChain).Return(false, nil).Times(1)
	mockIptables.EXPECT().NewChain(table, trackerChain).Return(nil).Times(1)
	mockIptables.EXPECT().Insert(table, inputChain, 1, firewall.jumpRuleSpec).Return(nil).Times(1)

	err := create(mockIptables, firewall.jumpRuleSpec)
	require.NoError(t, err)
}

func Test_newFW(t *testing.T) {
	ipv4, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	require.NoError(t, err)
	fw := newFW(ipv4, "192.168.0.147")
	require.NotNil(t, fw)
}

func Test_getLocalIPString(t *testing.T) {
	actual1 := getLocalIPString(nil)
	assert.Equal(t, "", actual1)

	localIPString := "192.44.55.66"
	localIP := net.ParseIP(localIPString)
	actual2 := getLocalIPString(localIP)
	assert.Equal(t, localIPString, actual2)
}
