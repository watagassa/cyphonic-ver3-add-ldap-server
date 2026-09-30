//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/psd/$GOPACKAGE
package repository

import (
	"context"
	"net/netip"

	"github.com/Pluslab/cyphonic/psd/entity"
)

type PacketHandler interface {
	Unmarshal(context.Context, []byte, int) (*entity.ProvisionRequest, error)
	GenerateProvisionResponse() *entity.ProvisionResponse
	SerializeBaseHeader(*entity.ProvisionResponse, entity.BaseHeader, entity.StatusClass)
	SerializeVirtualIPAddress(*entity.ProvisionResponse, netip.Prefix, netip.Prefix)
	SerializeL2Flag(*entity.ProvisionResponse)
	SerializeFQDN(*entity.ProvisionResponse, string)
	ChangeStatus(*entity.ProvisionResponse, entity.StatusClass)
}
