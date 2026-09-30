//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/fsd/$GOPACKAGE
package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/fsd/entity"
)

type SQLHandler interface {
	DeleteNodeInformation(context.Context, entity.FinalizationRequest) error
}
