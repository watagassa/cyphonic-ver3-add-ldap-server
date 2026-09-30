package entity

import (
	"fmt"
	"net"
	"net/netip"
	"slices"
	"sync"
)

type NSAddress struct {
	IPv4Addr  netip.Addr
	IPv6Addr  netip.Addr
	Port      uint16
	CommonKey []byte
	sync.RWMutex
}

// TypeNSIPVersion is defined to identify the IP version of an NS.
type TypeNSIPVersion uint8

const (
	TypeNSIPNone TypeNSIPVersion = iota
	TypeNSIPVersion4
	TypeNSIPVersion6
	TypeNSDualStackNetwork
)

func (a *NSAddress) Set(ipv4Addr, ipv6Addr netip.Addr, port uint16, commonKey []byte) {
	a.Lock()
	defer a.Unlock()
	a.IPv4Addr = ipv4Addr
	a.IPv6Addr = ipv6Addr
	a.Port = port
	a.CommonKey = slices.Clone(commonKey)
}

func (a *NSAddress) Reset() {
	a.Lock()
	defer a.Unlock()
	a.IPv4Addr = netip.Addr{}
	a.IPv6Addr = netip.Addr{}
	a.Port = 0
	a.CommonKey = nil
}

func (a *NSAddress) IPVersion() TypeNSIPVersion {
	a.RLock()
	defer a.RUnlock()

	return a.ipVersionLocked()
}

// ipVersionLocked returns the IP version of the NS address.
// A zero-filled address (0.0.0.0 / ::) means "not present" by protocol convention.
// The caller must hold the read or write lock.
func (a *NSAddress) ipVersionLocked() TypeNSIPVersion {
	v4 := a.IPv4Addr.IsValid() && !a.IPv4Addr.IsUnspecified()
	v6 := a.IPv6Addr.IsValid() && !a.IPv6Addr.IsUnspecified()

	switch {
	case v4 && v6:
		return TypeNSDualStackNetwork
	case v4:
		return TypeNSIPVersion4
	case v6:
		return TypeNSIPVersion6
	default:
		return TypeNSIPNone
	}
}

// Addr returns the appropriate NS address based on the node's local IP version.
func (a *NSAddress) Addr(localIPVersion TypeLocalIPVersion) (netip.Addr, error) {
	a.RLock()
	defer a.RUnlock()

	return a.addrLocked(localIPVersion)
}

// addrLocked returns the appropriate NS address based on the node's local IP version.
// The caller must hold the read or write lock.
func (a *NSAddress) addrLocked(localIPVersion TypeLocalIPVersion) (netip.Addr, error) {
	nsIPVersion := a.ipVersionLocked()
	switch localIPVersion {
	case TypeLocalDualStackNetwork:
		switch nsIPVersion {
		case TypeNSIPVersion4:
			return a.IPv4Addr, nil
		case TypeNSIPVersion6, TypeNSDualStackNetwork:
			return a.IPv6Addr, nil
		default:
			return netip.Addr{}, fmt.Errorf("unknown NS IP version: %d", nsIPVersion)
		}

	case TypeLocalIPVersion4:
		switch nsIPVersion {
		case TypeNSIPVersion4, TypeNSDualStackNetwork:
			return a.IPv4Addr, nil
		default:
			return netip.Addr{}, fmt.Errorf("unknown NS IP version: %d", nsIPVersion)
		}

	case TypeLocalIPVersion6:
		switch nsIPVersion {
		case TypeNSIPVersion6, TypeNSDualStackNetwork:
			return a.IPv6Addr, nil
		default:
			return netip.Addr{}, fmt.Errorf("unknown NS IP version: %d", nsIPVersion)
		}
	default:
		return netip.Addr{}, fmt.Errorf("unknown local IP version: %d", localIPVersion)
	}
}

// UDPAddr returns a net.UDPAddr based on the NS address and the node's local IP version.
func (a *NSAddress) UDPAddr(localIPVersion TypeLocalIPVersion) (net.UDPAddr, error) {
	a.RLock()
	defer a.RUnlock()

	nsAddr, err := a.addrLocked(localIPVersion)
	if err != nil {
		return net.UDPAddr{}, err
	}

	return net.UDPAddr{
		IP:   nsAddr.AsSlice(),
		Port: int(a.Port),
	}, nil

}

// IsNSAddr reports whether the given address matches either NS address.
func (a *NSAddress) IsNSAddr(addr netip.Addr) bool {
	a.RLock()
	defer a.RUnlock()

	return a.IPv4Addr.Unmap() == addr.Unmap() || a.IPv6Addr.Unmap() == addr.Unmap()
}
