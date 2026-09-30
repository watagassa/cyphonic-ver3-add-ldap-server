package repository

import (
	"net"
)

// PathCache is an interface that defines methods for managing path information in a cache.
type PathCache interface {
	GetTunnelKey(pathID string) (key []byte, keyCipherType uint16, keyLength uint16, err error)

	GetResponderNodeAddr(pathID string) (ipVer int, ipAddr net.IP, ipPort uint16, err error)
	GetInitiatorNodeAddr(pathID string) (ipVer int, ipAddr net.IP, ipPort uint16, err error)

	UpdateResponderNodeAddr(pathID string, ip net.IP, port uint16) error
	UpdateInitiatorNodeAddr(pathID string, ip net.IP, port uint16) error

	Update(pathID string)
}
