// Package infrastructure contains the UDP connection.
package infrastructure

import (
	"net"

	"github.com/Pluslab/cyphonic/u-trsd/usecase/repository"
)

var _ repository.UDPConn = (*udpConn)(nil)

type udpConn struct {
	conn *net.UDPConn
}

// NewUDPConn creates a new UDP connection on the specified port.
func NewUDPConn(port string) (repository.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &udpConn{conn: conn}, nil
}

func (u *udpConn) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	n, addr, err := u.conn.ReadFromUDP(b)
	if err != nil {
		return 0, nil, err
	}

	return n, addr, nil
}

func (u *udpConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	n, err := u.conn.WriteToUDP(b, addr)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (u *udpConn) Close() error {
	return u.conn.Close()
}
