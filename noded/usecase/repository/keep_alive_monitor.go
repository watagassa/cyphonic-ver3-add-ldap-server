package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/noded/entity"
)

type KeepAliveMonitor interface {
	Run(ctx context.Context,
		keepAliveQueue chan<- entity.OutPacketQueue,
		keepAliveAckQueue <-chan struct{},
		reRegistrationQueue chan<- entity.OutPacketQueue,
	) error
}
