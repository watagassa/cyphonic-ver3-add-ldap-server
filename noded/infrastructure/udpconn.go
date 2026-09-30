package infrastructure

import (
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ repository.UDPConn = (*udpConn)(nil)

type udpConn struct {
	conn *net.UDPConn
}

func NewUDPConn(port string) (repository.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen UDP: %w", err)
	}

	return &udpConn{conn: conn}, nil
}

func (u *udpConn) LocalAddr() string {
	return u.conn.LocalAddr().String()
}

func (u *udpConn) RemoteAddr() string {
	return u.conn.RemoteAddr().String()
}

func (u *udpConn) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	n, addr, err := u.conn.ReadFromUDP(b)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read from UDP: %w", err)
	}

	return n, addr, nil
}

func (u *udpConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	n, err := u.conn.WriteToUDP(b, addr)
	if err != nil {
		return 0, fmt.Errorf("failed to write to UDP: %w", err)
	}

	return n, nil
}

func (u *udpConn) Close() error {
	if err := u.conn.Close(); err != nil {
		return fmt.Errorf("failed to close UDP connection: %w", err)
	}

	return nil
}
