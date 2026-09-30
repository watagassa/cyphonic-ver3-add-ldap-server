package repository

import (
	"context"
	"syscall"
)

type NetlinkListener interface {
	WatchIPAddrChange(ctx context.Context, ipAddrChangedCh chan<- bool) error
	ReadMsgs() ([]syscall.NetlinkMessage, error)
}
