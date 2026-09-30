//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/cmsd/$GOPACKAGE
package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/cmsd/entity"
)

type PacketHandler interface {
	Unmarshal(context.Context, []byte, int) (*entity.ConnectionRequest, error)
	GenerateConnectionResponse() *entity.ConnectionResponse
	SerializeBaseHeader(*entity.ConnectionResponse, entity.BaseHeader, entity.StatusClass)
	SerializeNSInformation(*entity.ConnectionResponse, *entity.NotificationServiceInfomation)
	SerializeCommonKey(*entity.ConnectionResponse, *entity.CommonKey)
	ChangeStatus(*entity.ConnectionResponse, entity.StatusClass)
}
