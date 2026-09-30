package repository

import (
	"net"

	"github.com/Pluslab/cyphonic/u-trsd/entity"
)

// SQLHandler is an interface that defines methods for interacting with the SQL database.
type SQLHandler interface {
	Close() error
	CreateTunnelRelayService(entity.TunnelRelayService) error
	DeleteTunnelRelayService(trsID string) error
	GetPathInformationByPathID(pathID string) (entity.PathInformation, error)
	SetResponderNodeAddress(pathID string, address net.IP, port uint16) error
	SetInitiatorNodeAddress(pathID string, address net.IP, port uint16) error
}
