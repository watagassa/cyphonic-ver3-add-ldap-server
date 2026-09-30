//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/fsd/$GOPACKAGE
package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/fsd/entity"
)

type PacketHandler interface {
	Unmarshal(context.Context, []byte, int) (*entity.FinalizationRequest, bool, error)
	GenerateFinalizationResponse(baseHeader entity.BaseHeader) (*entity.FinalizationResponse, error)
}
