//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/asd/$GOPACKAGE
package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/asd/entity"
)

type SQLHandler interface {
	GetDevice(context.Context, entity.LoginRequest, []byte, []byte) (entity.Device, entity.StatusClass, error)
	GetVirtualIPAddress(entity.Device) (entity.VirtualIPAddress, error)
	SetNodeInformation(context.Context, entity.Device, entity.VirtualIPAddress, *entity.LoginRequest, []byte) error
}
