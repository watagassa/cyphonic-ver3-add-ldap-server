package repository

import (
	"context"
	"net/netip"

	"github.com/Pluslab/cyphonic/noded/entity"
)

type RouteDirectionCache interface {
	Run(ctx context.Context) error
	Get(key string) (routeDirection entity.RouteDirection, shouldUpdatePeerAddr bool, peerAddr netip.AddrPort, err error)
	Put(key string, routeDirection *entity.RouteDirection, expires int64) error
	PutPeerAddr(key string, addr netip.AddrPort, expires int64) error
}
