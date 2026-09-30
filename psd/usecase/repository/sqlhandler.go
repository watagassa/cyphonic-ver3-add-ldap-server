//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/psd/$GOPACKAGE
package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/psd/entity"
)

type SQLHandler interface {
	GetNodeInformation(context.Context, entity.ID) (entity.NodeInformation, error)
	GetAliasFQDN(context.Context, string, string) (entity.FQDNAlias, error)
	UpdateNodeInformation(context.Context, entity.NodeInformation) error
}
