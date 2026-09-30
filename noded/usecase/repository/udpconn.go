//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/noded/$GOPACKAGE
package repository

import "net"

type UDPConn interface {
	LocalAddr() string
	RemoteAddr() string
	ReadFromUDP(buffer []byte) (int, *net.UDPAddr, error)
	WriteToUDP(buffer []byte, addr *net.UDPAddr) (int, error)
	Close() error
}
