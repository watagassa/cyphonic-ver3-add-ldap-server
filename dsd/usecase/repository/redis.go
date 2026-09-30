//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/dsd/$GOPACKAGE
package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/dsd/entity"
)

type RedisClient interface {
	Close() error
	CreateOrUpdatePathInformation(ctx context.Context, pathID *entity.ID,
		initiatorNodeAddress, responderNodeAddress *entity.NodeAddress, tunnelKey []byte) error
	GetInitiatorNodeIDByPathID(ctx context.Context, pathID *entity.ID) (entity.ID, error)
	SetInitiatorNodeIDAndResponderFQDN(ctx context.Context, pathID *entity.ID, initiatorNodeID entity.ID, responderFQDN string) error
}
