package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/noded/entity"
)

type TunHandler interface {
	Read(ctx context.Context, outRawMsgQueue chan<- entity.OutboundMsgQueue) error
	Write(ctx context.Context, inRawMsgQueue <-chan *entity.InboundMsgQueue) error
}
