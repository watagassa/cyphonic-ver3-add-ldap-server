//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/dsd/$GOPACKAGE
package repository

import "github.com/Pluslab/cyphonic/dsd/entity"

type SQLHandler interface {
	Close() error
	GetNodeInformationAndNodeAddressByNodeID(nodeID string) (entity.NodeInformation, entity.NodeAddress, error)
	GetNodeInformationAndNodeAddressByFQDN(fqdn string) (entity.NodeInformation, entity.NodeAddress, error)
	GetNotificationServiceByNSID(nsid string) (entity.NotificationService, error)
	GetRandomUDPTunnelRelayService() (entity.TunnelRelayService, error)
	SetPathInformation(pathID, tunnelKey []byte, initiatorNodeAddress, responderNodeAddress entity.NodeAddress) error
}
