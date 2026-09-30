package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/noded/entity"
)

type TunDnsHandler interface {
	Read(ctx context.Context, dnsQueue chan<- entity.DNSQueue) error
	Write(ctx context.Context,
		dnsAnswerQueue <-chan entity.DNSAnswerQueue,
		cachedDnsAnswerQueue <-chan entity.DNSAnswerQueue) error
}
