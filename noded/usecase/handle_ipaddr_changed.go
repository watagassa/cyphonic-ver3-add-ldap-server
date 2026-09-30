package usecase

import (
	"context"
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
)

func (h *handler) HandleIPAddrChanged(ctx context.Context, ipAddrChangedCh <-chan bool, reRegistrationQueue chan<- entity.OutPacketQueue) error {
	config.LogDebug("starting handle ip addr changed function")
	defer config.LogDebug("finished handle ip addr changed function")

	for {
		var (
			buf         []byte
			soc         entity.Socket
			sendAddress *net.UDPAddr
			err         error
		)
		select {
		case <-ctx.Done():
			return nil
		case <-ipAddrChangedCh:
			buf, soc, sendAddress, err = h.generateReRegistration()
		}

		if err != nil {
			config.LogErr("Failed to generate Re Registration Request", "error", err)
			continue
		}

		if buf != nil && sendAddress != nil {
			reRegistrationQueue <- entity.OutPacketQueue{
				Buffer: buf,
				Addr:   sendAddress,
				Soc:    soc,
			}
		}

	}
}

// TODO: daemon(nsService)の関数と重複してるの直す
func (h *handler) generateReRegistration() ([]byte, entity.Socket, *net.UDPAddr, error) {
	// TODO: ソケットの優先度変動処理実装
	soc, err := h.socketStore.HighestPrioritySocket()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get highest priority socket: %w", err)
	}

	nodeIPv4Address := soc.LocalIPv4Addr()
	nodeIPv6Address := soc.LocalIPv6Addr()
	registrationRequest := entity.GenerateRegistrationRequest(h.baseHeader, nodeIPv4Address, nodeIPv6Address, soc.InterfaceName())

	buffer, err := registrationRequest.Marshal()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal re registration request: %w", err)
	}

	// TODO: 送信前にcommonKeyで暗号化しないといけない
	sendAddress, err := h.nsAddress.UDPAddr(h.localIPVersion)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get NS UDP address: %w", err)
	}

	config.LogDebug("Generated Re Registration Request To NS", "dst", sendAddress.String())

	return buffer, soc, &sendAddress, nil
}
