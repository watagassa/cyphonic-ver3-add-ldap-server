package infrastructure

import (
	"context"
	"fmt"
	"syscall"

	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ repository.NetlinkListener = (*netlinkListener)(nil)

type netlinkListener struct {
	fd int
	sa *syscall.SockaddrNetlink
}

func NewNetlinkListener() (repository.NetlinkListener, error) {
	groups := (1 << (syscall.RTNLGRP_LINK - 1)) |
		(1 << (syscall.RTNLGRP_IPV4_IFADDR - 1)) |
		(1 << (syscall.RTNLGRP_IPV6_IFADDR - 1))

	s, err := syscall.Socket(
		syscall.AF_NETLINK,
		syscall.SOCK_DGRAM,
		syscall.NETLINK_ROUTE,
	)

	if err != nil {
		return nil, fmt.Errorf("netlink socket: %w", err)
	}

	saddr := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Pid:    uint32(0),
		Groups: uint32(groups),
	}

	err = syscall.Bind(s, saddr)
	if err != nil {
		return nil, fmt.Errorf("netlink bind: %w", err)
	}

	return &netlinkListener{
		fd: s,
		sa: saddr,
	}, nil
}

func (l *netlinkListener) WatchIPAddrChange(ctx context.Context, ipAddrChangedCh chan<- bool) error {
	config.LogDebug("Starting watch ip addr change function")
	defer config.LogDebug("Finished watch ip addr change function")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msgs, err := l.ReadMsgs()
			if err != nil {
				return fmt.Errorf("netlink read: %w", err)
			}

			for _, m := range msgs {
				if isNewAddr(&m) {
					ipAddrChangedCh <- true
				}

				if isDelAddr(&m) {
					ipAddrChangedCh <- true
				}
			}
		}
	}
}

// ReadMsgs reads file descriptor's messages.
func (l *netlinkListener) ReadMsgs() ([]syscall.NetlinkMessage, error) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("netlink read msgs: %v\n", err)
		}
	}()

	pkt := make([]byte, 2048)

	n, err := syscall.Read(l.fd, pkt)
	if err != nil {
		return nil, fmt.Errorf("netlink read: %w", err)
	}

	msgs, err := syscall.ParseNetlinkMessage(pkt[:n])
	if err != nil {
		return nil, fmt.Errorf("netlink parse: %w", err)
	}

	return msgs, nil
}

// isNewAddr monitors if an IP address has been added to the network interface.
func isNewAddr(msg *syscall.NetlinkMessage) bool {
	return msg.Header.Type == syscall.RTM_NEWADDR
}

// isDelAddr monitors if an IP address has been deleted to the network interface.
func isDelAddr(msg *syscall.NetlinkMessage) bool {
	return msg.Header.Type == syscall.RTM_DELADDR
}
