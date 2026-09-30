//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/u-nsd/$GOPACKAGE
package repository

import (
	"net"

	"github.com/Pluslab/cyphonic/u-nsd/entity"
)

type SQLHandler interface {
	Close() error
	GetNodeInformationByNodeID(nodeID entity.ID) (*entity.NodeInformation, error)
	SaveInitiatorNodeAddress(interfaceName string,
		registrationRequest *entity.RegistrationRequest, natIPv4, natIPv6 net.IP, natPort int) (*entity.NodeAddress, error)
	CreateSelfNotificationService() error
	DeleteNodeAddressesBySelfNSID() error
	DeleteSelfNotificationService() error
}
