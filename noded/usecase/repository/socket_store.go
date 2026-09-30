//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/noded/$GOPACKAGE
package repository

import "github.com/Pluslab/cyphonic/noded/entity"

type SocketStore interface {
	Values() map[entity.InterfaceName]entity.Socket
	HighestPrioritySocket() (entity.Socket, error)
	ExpireDeadlines() error
	ClearDeadlines() error
	SwitchActiveConn(localIPVersion entity.TypeLocalIPVersion)
}
