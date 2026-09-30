package entity

import (
	"net"
	"time"
)

type NodeAddress struct {
	NodeID        []byte     `gorm:"column:node_id"`
	InterfaceName string     `gorm:"column:interface_id"`
	RealIPv4      []byte     `gorm:"column:real_ipv4"`
	RealIPv6      []byte     `gorm:"column:real_ipv6"`
	NATIPv4       []byte     `gorm:"column:nat_ipv4"`
	NATIPv6       []byte     `gorm:"column:nat_ipv6"`
	NATPort       int        `gorm:"column:nat_port"`
	NSID          string     `gorm:"column:ns_id"`
	CreatedAt     *time.Time `gorm:"column:created_at;gorm:<-:create"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;gorm:<-:update"`
}

// NetworkType defines the network type.
type NetworkType string

// NetworkType known values.
const (
	NetworkTypeIPv4      NetworkType = "ipv4"
	NetworkTypeIPv6      NetworkType = "ipv6"
	NetworkTypeDualStack NetworkType = "dualstack"
	NetworkTypeNeedTRS   NetworkType = "needTRS"
)

func (na NodeAddress) ShouldRelay(responder *NodeAddress) bool {
	networkType := mostSuitableIPVersion(na, *responder)

	switch networkType {
	case NetworkTypeIPv4:
		return manageFlag(na.RealIPv4, responder.RealIPv4, na.NATIPv4, responder.NATIPv4)
	case NetworkTypeIPv6:
		return manageFlag(na.RealIPv6, responder.RealIPv6, na.NATIPv6, responder.NATIPv6)
	case NetworkTypeDualStack:
		return manageFlag(na.RealIPv4, responder.RealIPv4, na.NATIPv4, responder.NATIPv4) && manageFlag(na.RealIPv6, responder.RealIPv6, na.NATIPv6, responder.NATIPv6)
	case NetworkTypeNeedTRS:
		return true
	}

	return false
}

func (na NodeAddress) ShouldHolePunching(responder *NodeAddress) bool {
	networkType := mostSuitableIPVersion(na, *responder)

	switch networkType {
	case NetworkTypeIPv4:
		return hasNATOnlyCN(na.RealIPv4, responder.RealIPv4, na.NATIPv4, responder.NATIPv4)
	case NetworkTypeIPv6:
		return hasNATOnlyCN(na.RealIPv6, responder.RealIPv6, na.NATIPv6, responder.NATIPv6)
	case NetworkTypeDualStack:
		// TODO: DualStackの場合、条件によってIPv4とIPv6を使い分けるか要検討
		return hasNATOnlyCN(na.RealIPv6, responder.RealIPv6, na.NATIPv6, responder.NATIPv6)
	case NetworkTypeNeedTRS:
		return false
	}

	return false
}

func manageFlag(mnRealIP, cnRealIP, mnNATIP, cnNATIP []byte) bool {
	mnNATFlag := hasNAT(mnRealIP, mnNATIP)
	cnNATFlag := hasNAT(cnRealIP, cnNATIP)
	sameNATFlag := hasSameNAT(mnNATIP, cnNATIP)

	if sameNATFlag || (mnNATFlag && cnNATFlag) {
		return true
	}

	return false
}

func hasNAT(realIP, natIP []byte) bool {
	return !net.IP(realIP).Equal(net.IP(natIP))
}

func hasSameNAT(mnNATIP, cnNATIP []byte) bool {
	return net.IP(mnNATIP).Equal(net.IP(cnNATIP))
}

func hasNATOnlyCN(mnRealIP, cnRealIP, mnNATIP, cnNATIP []byte) bool {
	mnNATFlag := hasNAT(mnRealIP, mnNATIP)
	cnNATFlag := hasNAT(cnRealIP, cnNATIP)

	if !mnNATFlag && cnNATFlag {
		return true
	}

	return false
}

// mostSuitableIPVersion determines if TRS should be used.
func mostSuitableIPVersion(mnAddr, cnAddr NodeAddress) NetworkType {
	mnNetworkType := analyzeNetworkType(mnAddr)
	cnNetworkType := analyzeNetworkType(cnAddr)

	switch {
	case mnNetworkType == NetworkTypeIPv4 && cnNetworkType == NetworkTypeIPv4:
		return NetworkTypeIPv4
	case mnNetworkType == NetworkTypeIPv6 && cnNetworkType == NetworkTypeIPv6:
		return NetworkTypeIPv6
	case mnNetworkType == NetworkTypeDualStack && cnNetworkType == NetworkTypeDualStack:
		return NetworkTypeDualStack
	}

	return NetworkTypeNeedTRS
}

func analyzeNetworkType(addr NodeAddress) NetworkType {
	if net.IPv4(0, 0, 0, 0).Equal(net.IP(addr.NATIPv4)) {
		return NetworkTypeIPv6
	} else if net.IPv6zero.Equal(net.IP(addr.NATIPv6)) {
		return NetworkTypeIPv4
	}

	return NetworkTypeDualStack
}
