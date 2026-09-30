package repository

import "context"

type DNSCache interface {
	Run(ctx context.Context) error
	Get(key uint32) (port, transactionID uint16)
	Put(key uint32, port, transactionID uint16, expires int64)
}
