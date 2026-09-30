package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ repository.TunHandler = (*tunHandler)(nil)

// IPv4 offset length.
const (
	IPv4offsetTotalLength = 2                           // IPv4offsetPayloadLength is IPv6 offset payload length.
	IPv4offsetSrc         = 12                          // IPv4offsetSrc is IPv6 offset src length.
	IPv4offsetDst         = IPv4offsetSrc + net.IPv4len // IPv4offsetDst is IPv6 offset dst length.
)

// IPv6 offset length.
const (
	IPv6offsetPayloadLength = 4                           // IPv6offsetPayloadLength is IPv6 offset payload length.
	IPv6offsetSrc           = 8                           // IPv6offsetSrc is IPv6 offset src length.
	IPv6offsetDst           = IPv6offsetSrc + net.IPv6len // IPv6offsetDst is IPv6 offset dst length.
)

type tunHandler struct {
	isVirtualIPv6    bool
	tunInterfaceName string
	tunInterface     *entity.TunInterface
	v4Address        string
	v6Address        string
	v4Prefix         int
	v6Prefix         int
}

func NewTunHandler(virtualIPType int, tunInterfaceName string, v4Address, v6Address string, v4Prefix, v6Prefix int) (repository.TunHandler, error) {
	var (
		isVirtualIPv6 bool
		tun           *entity.TunInterface
		err           error
	)

	switch virtualIPType {
	case 4:
		tun, err = entity.NewTunInterface(tunInterfaceName, v4Address, v4Prefix)
		isVirtualIPv6 = false
	case 6:
		tun, err = entity.NewTunInterface(tunInterfaceName, v6Address, v6Prefix)
		isVirtualIPv6 = true
	default:
		return nil, fmt.Errorf("unknown ip version: %d", virtualIPType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create TUN interface: %w", err)
	}

	if err = tun.Up(); err != nil {
		_ = tun.Close()
		return nil, fmt.Errorf("failed to up tun interface: %w", err)
	}

	config.LogDebug("Start virtual I/F",
		"tun_name", tun.Name(),
		"vip_addr", tun.Address())

	return &tunHandler{
		tunInterfaceName: tunInterfaceName,
		tunInterface:     tun,
		isVirtualIPv6:    isVirtualIPv6,
		v4Address:        v4Address,
		v6Address:        v6Address,
		v4Prefix:         v4Prefix,
		v6Prefix:         v6Prefix,
	}, nil
}

func (t *tunHandler) Read(ctx context.Context, outRawMsgQueue chan<- entity.OutboundMsgQueue) error {
	config.LogDebug("Starting to read data from TUN interface")
	defer config.LogDebug("Finished reading data from TUN interface")

	stop := context.AfterFunc(ctx, func() {
		_ = t.tunInterface.Close()
	})
	defer stop()

	for {
		buf := new([65535]byte) // TODO: Poolに置換
		size, err := t.tunInterface.Read(buf[:])
		if err != nil {
			if errors.Is(err, os.ErrClosed) {
				config.LogDebug("TUN interface closed, stopping data reception")
				return nil
			}
			config.LogErr("Failed to read data from TUN interface", "error", err)
			continue
		}

		if size == 0 {
			config.LogErr("Received packet is too small", "error", size)
			continue
		}

		// get the destination IP address
		var dst []byte
		if !t.isVirtualIPv6 {
			dst = buf[IPv4offsetDst : IPv4offsetDst+net.IPv4len]
		} else {
			dst = buf[IPv6offsetDst : IPv6offsetDst+net.IPv6len]
		}
		// convert to netip.Addr type
		dstAddr, ok := netip.AddrFromSlice(dst)
		if !ok {
			config.LogErr("Failed to convert dst to netip.Addr", "dst", dst)
			continue
		}

		outRawMsgQueue <- entity.OutboundMsgQueue{
			Data: buf[:size],
			Dst:  &dstAddr,
		}
	}
}

func (t *tunHandler) Write(ctx context.Context, inRawMsgQueue <-chan *entity.InboundMsgQueue) error {
	config.LogDebug("Starting to write data to TUN interface")
	defer config.LogDebug("Finished writing data to TUN interface")

	stop := context.AfterFunc(ctx, func() {
		_ = t.tunInterface.Close()
	})
	defer stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case inRawMsg := <-inRawMsgQueue:
			inRawMsg.Lock()
			if inRawMsg.Data == nil {
				continue
			}

			_, err := t.tunInterface.Write(inRawMsg.Data)
			if err != nil {
				if errors.Is(err, os.ErrClosed) {
					config.LogDebug("TUN interface closed, stopping data transmission")
					return nil
				}
				config.LogErr("Failed to write to TUN interface", "error", err)
				continue
			}
		}
	}
}
