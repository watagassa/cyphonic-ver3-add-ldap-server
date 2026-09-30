package entity

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"syscall"
	"time"
)

type Socket interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	Close() error

	SwitchActiveConn(localIPVersion TypeLocalIPVersion)
	SetReadDeadline(t time.Time) error

	InterfaceName() string
	LocalAddr() string
	LocalIPv4String() string
	LocalIPv6String() string
	LocalIPv4Addr() netip.Addr
	LocalIPv6Addr() netip.Addr
	IPVersionString() string
	IPVersion() TypeLocalIPVersion
	Priority() int
}

var _ Socket = (*socket)(nil)

type socket struct {
	// conn is the active connection selected based on the current IP version.
	conn          *net.UDPConn
	connV4        *net.UDPConn
	connV6        *net.UDPConn
	ipVersion     TypeLocalIPVersion
	interfaceName string
	priority      int
}

type InterfaceName string

func NewSocket(interfaceName string, ipAddr4, ipAddr6 netip.Addr, port string, priority int) (Socket, error) {
	if !ipAddr4.IsValid() && !ipAddr6.IsValid() {
		return nil, fmt.Errorf("both IPv4 and IPv6 addresses are invalid")
	}

	var connV4, connV6, conn *net.UDPConn
	var err error

	if ipAddr4.IsValid() {
		connV4, err = newUDPConn(interfaceName, ipAddr4, port)
		if err != nil {
			return nil, fmt.Errorf("failed to create IPv4 UDP connection: %w", err)
		}
	}

	if ipAddr6.IsValid() {
		connV6, err = newUDPConn(interfaceName, ipAddr6, port)
		if err != nil {
			if connV4 != nil {
				connV4.Close()
			}
			return nil, fmt.Errorf("failed to create IPv6 UDP connection: %w", err)
		}
	}

	var ipVersion TypeLocalIPVersion
	if (connV4 != nil) && (connV6 != nil) {
		ipVersion = TypeLocalDualStackNetwork
		conn = connV6
	} else if connV4 != nil {
		ipVersion = TypeLocalIPVersion4
		conn = connV4
	} else if connV6 != nil {
		ipVersion = TypeLocalIPVersion6
		conn = connV6
	} else {
		ipVersion = TypeLocalIPNone
	}

	return &socket{
		connV4:        connV4,
		connV6:        connV6,
		conn:          conn,
		ipVersion:     ipVersion,
		interfaceName: interfaceName,
		priority:      priority,
	}, nil
}

func newUDPConn(interfaceName string, ipAddr netip.Addr, port string) (*net.UDPConn, error) {
	var lAddr string
	if ipAddr.Is6() {
		lAddr = "[" + ipAddr.String() + "]:" + port
	} else {
		lAddr = ipAddr.String() + ":" + port
	}
	addr, err := net.ResolveUDPAddr("udp", lAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen UDP: %w", err)
	}

	sysConn, err := conn.SyscallConn()
	if err != nil {
		fmt.Printf("Unable to get conn's netFD: %v\n", err)
		conn.Close()
		os.Exit(1)
	}

	// run setsockopt to set the interface for unicast packets.
	var setsockoptErr error
	err = sysConn.Control(func(fd uintptr) {
		setsockoptErr = syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, interfaceName)
	})

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("sysConn.Control() failed: %w", err)
	}
	if setsockoptErr != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind to device %s: %w", interfaceName, setsockoptErr)
	}

	return conn, nil
}

func (s *socket) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	return s.conn.ReadFromUDP(b)
}

func (s *socket) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	return s.conn.WriteToUDP(b, addr)
}

func (s *socket) LocalAddr() string {
	if s.conn == nil {
		return ""
	}
	return s.conn.LocalAddr().String()
}

func (s *socket) LocalIPv4String() string {
	if s.connV4 == nil {
		return ""
	}
	return s.connV4.LocalAddr().String()
}

func (s *socket) LocalIPv6String() string {
	if s.connV6 == nil {
		return ""
	}
	return s.connV6.LocalAddr().String()
}

// LocalIPv4Addr returns the local IPv4 address, or an empty netip.Addr if the IPv4 connection does not exist.
func (s *socket) LocalIPv4Addr() netip.Addr {
	if s.connV4 == nil {
		return netip.Addr{}
	}

	localAddr, ok := s.connV4.LocalAddr().(*net.UDPAddr)
	if !ok {
		return netip.Addr{}
	}
	return localAddr.AddrPort().Addr()
}

// LocalIPv6Addr returns the local IPv6 address, or an empty netip.Addr if the IPv6 connection does not exist.
func (s *socket) LocalIPv6Addr() netip.Addr {
	if s.connV6 == nil {
		return netip.Addr{}
	}

	localAddr, ok := s.connV6.LocalAddr().(*net.UDPAddr)
	if !ok {
		return netip.Addr{}
	}

	return localAddr.AddrPort().Addr()
}

func (s *socket) IPVersion() TypeLocalIPVersion {
	return s.ipVersion
}

func (s *socket) IPVersionString() string {
	switch s.ipVersion {
	case TypeLocalIPVersion4:
		return "IPv4"
	case TypeLocalIPVersion6:
		return "IPv6"
	case TypeLocalDualStackNetwork:
		return "DualStack"
	default:
		return "None"
	}
}

func (s *socket) Close() error {
	if err := s.connV4.Close(); err != nil {
		return fmt.Errorf("failed to close IPv4 connection: %w", err)
	}
	if err := s.connV6.Close(); err != nil {
		return fmt.Errorf("failed to close IPv6 connection: %w", err)
	}

	return nil
}

func (s *socket) Priority() int {
	return s.priority
}

func (s *socket) InterfaceName() string {
	return s.interfaceName
}

func (s *socket) SwitchActiveConn(localIPVersion TypeLocalIPVersion) {
	switch localIPVersion {
	case TypeLocalDualStackNetwork, TypeLocalIPVersion6:
		s.conn = s.connV6
	case TypeLocalIPVersion4:
		s.conn = s.connV4
	}
}

func (s *socket) SetReadDeadline(t time.Time) error {
	if conn := s.connV4; conn != nil {
		err := conn.SetReadDeadline(t)
		if err != nil {
			return fmt.Errorf("failed to clear read deadline for IPv4 connection: %w", err)
		}
	}

	if conn := s.connV6; conn != nil {
		err := conn.SetReadDeadline(t)
		if err != nil {
			return fmt.Errorf("failed to clear read deadline for IPv6 connection: %w", err)
		}
	}
	return nil
}
