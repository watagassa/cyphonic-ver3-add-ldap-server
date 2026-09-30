//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/fsd/$GOPACKAGE
package repository

import "net"

type TLSListener interface {
	Accept() (net.Conn, error)
	Close() error
}
