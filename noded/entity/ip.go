package entity

import (
	"fmt"
	"net"
	"net/netip"
)

// TypeLocalIPVersion defines to identify the local IP version of noded.
type TypeLocalIPVersion uint8

// TypeLocalIPVersion knows values.
const (
	TypeLocalIPNone TypeLocalIPVersion = iota
	TypeLocalIPVersion4
	TypeLocalIPVersion6
	TypeLocalDualStackNetwork
)

// VIPv4HostMatchPrefix VIPv6HostMatchPrefix Define an exact match prefix book that identifies unique hosts.
const (
	VIPv4HostMatchPrefix = 32
	VIPv6HostMatchPrefix = 128
)

func ZeroAddr4() netip.Addr {
	ipv4zero, _ := netip.AddrFromSlice(net.IPv4zero)

	return ipv4zero.Unmap()
}

func ZeroAddr6() netip.Addr {
	ipv6zero, _ := netip.AddrFromSlice(net.IPv6zero)

	return ipv6zero
}

func GetLocalIP(match func(addr netip.Addr) bool) (netip.Addr, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return netip.Addr{}, fmt.Errorf("can't get network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		// skip down or loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue // skip interface if error occurs
		}

		for _, a := range addrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}

			addr, ok := netip.AddrFromSlice(ipNet.IP)
			if !ok || addr.IsLoopback() {
				continue
			}

			if match(addr.Unmap()) {
				return addr.Unmap(), nil
			}
		}
	}

	return netip.Addr{}, fmt.Errorf("no matching IP address found")
}

func DetermineLocalIPVersion(ipv4Address, ipv6Address netip.Addr) TypeLocalIPVersion {
	switch {
	case ipv4Address.IsValid() && ipv6Address.IsValid():
		return TypeLocalDualStackNetwork
	case ipv4Address.IsValid() && !ipv6Address.IsValid():
		return TypeLocalIPVersion4
	case !ipv4Address.IsValid() && ipv6Address.IsValid():
		return TypeLocalIPVersion6
	default:
		return TypeLocalIPNone
	}
}

// // invalidTUNIPv4 compares the IPv4 addresses received as arguments.
// // and, returns a bool.
// func invalidTUNIPv4(targetIP netip.Addr) bool {
// 	dnsTunIPv4 := netip.MustParseAddr(os.Getenv("TUN_DNS_INTERFACE_IPv4"))
// 	prefix := netip.MustParsePrefix(VIPv4Prefix)
// 	pf := netip.PrefixFrom(targetIP, VIPv4PrefixLen)

// 	if (targetIP != dnsTunIPv4) && !(prefix.Overlaps(pf)) {
// 		return true
// 	}

// 	return false
// }

// // invalidTUNIPv6 compares the IPv6 addresses received as arguments.
// // and, returns a bool.
// func invalidTUNIPv6(targetIP netip.Addr) bool {
// 	dnsTunIPv6 := netip.MustParseAddr(os.Getenv("TUN_DNS_INTERFACE_IPv6"))
// 	prefix := netip.MustParsePrefix(VIPv6Prefix)
// 	pf := netip.PrefixFrom(targetIP, VIPv6PrefixLen)

// 	if (targetIP != dnsTunIPv6) && !(prefix.Overlaps(pf)) {
// 		return true
// 	}

// 	return false
// }
