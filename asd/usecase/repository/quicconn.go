//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/asd/$GOPACKAGE
package repository

import (
	"context"
)

type QUICListener interface {
	Accept(context.Context) (ConnectionRepository, error)
	Close() error
}
