//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/fsd/$GOPACKAGE
package repository

import (
	"context"

	quic "github.com/quic-go/quic-go"
)

type QUICListener interface {
	Accept(context.Context) (quic.Connection, error)
	Close() error
}
