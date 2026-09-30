package infrastructure

import (
	"fmt"
	"net"
	"strconv"

	"github.com/Pluslab/cyphonic/u-nsd/usecase/repository"
)

type udpConn struct {
	conn   *net.UDPConn
	dsAddr *net.UDPAddr
}

func NewUDPConn(port, dsIPv4, dsIPv6, strDsPort string) (repository.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP port: %w", err)
	}

	dsPort, err := strconv.Atoi(strDsPort)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DS port to integer: %w", err)
	}
	dsAddr := &net.UDPAddr{
		IP:   net.ParseIP(dsIPv4),
		Port: dsPort,
	}

	return &udpConn{
		conn:   conn,
		dsAddr: dsAddr,
	}, nil
}

func (u *udpConn) GetDSAddr() *net.UDPAddr {
	return u.dsAddr
}

func (u *udpConn) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	n, addr, err := u.conn.ReadFromUDP(b)
	if err != nil {
		return n, addr, fmt.Errorf("failed to read from UDP: %w", err)
	}

	return n, addr, nil
}

func (u *udpConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	n, err := u.conn.WriteToUDP(b, addr)
	if err != nil {
		return n, fmt.Errorf("failed to write to UDP: %w", err)
	}

	return n, nil
}

func (u *udpConn) Close() error {
	if err := u.conn.Close(); err != nil {
		return fmt.Errorf("failed to close UDP connection: %w", err)
	}

	return nil
}
