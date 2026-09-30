package repository

import "net"

// UDPConn is an interface that defines methods for managing a UDP connection.
type UDPConn interface {
	ReadFromUDP([]byte) (int, *net.UDPAddr, error)
	WriteToUDP([]byte, *net.UDPAddr) (int, error)
	Close() error
}
