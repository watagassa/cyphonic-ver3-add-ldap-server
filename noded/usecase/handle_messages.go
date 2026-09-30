package usecase

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"golang.org/x/sync/errgroup"
)

func (h *handler) HandleInboundMessages(ctx context.Context,
	stagedInboundMsgQueue <-chan entity.StagedInboundMsgQueue) error {
	config.LogDebug("Starting decrypt inbound messages function")
	defer config.LogDebug("Finished decrypt inbound messages function")

	errGroup, ctx := errgroup.WithContext(ctx)

	for i := 0; i < numWorkers; i++ {
		errGroup.Go(func() error {
			for {
				var msg entity.StagedInboundMsgQueue
				select {
				case <-ctx.Done():
					return nil
				case msg = <-stagedInboundMsgQueue:
				}

				queuePtr := msg.QueuePtr
				buf, err := h.handleCapsuleMessage(msg.Packet)
				if err != nil {
					config.LogErr("Failed to handle capsule message", "error", err)
					queuePtr.Unlock()
					continue
				}

				queuePtr.Data = buf
				queuePtr.Unlock()
			}
		})
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("error in decrypt inbound messages workers: %w", err)
	}

	return nil
}

func (h *handler) HandleOutboundMessages(ctx context.Context,
	stagedOutboundMsgQueue <-chan entity.StagedOutboundMsgQueue) error {
	config.LogDebug("Starting handle outbound messages function")
	defer config.LogDebug("Finished handle outbound messages function")

	errGroup, ctx := errgroup.WithContext(ctx)

	for i := 0; i < numWorkers; i++ {
		errGroup.Go(func() error {
			for {
				var msg entity.StagedOutboundMsgQueue
				select {
				case <-ctx.Done():
					return nil
				case msg = <-stagedOutboundMsgQueue:
				}

				queuePtr := msg.QueuePtr
				buf, sendAddress, err := h.generateCapsuleMessage(msg.Data, *msg.Dst)

				if err != nil {
					config.LogErr("Failed to generate capsule message", "error", err)
					queuePtr.Unlock()
					continue
				}

				queuePtr.Buffer = buf
				queuePtr.Addr = sendAddress
				queuePtr.Unlock()
			}
		})
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("error in handle outbound messages workers: %w", err)
	}

	return nil
}

func (h *handler) handleCapsuleMessage(packet *entity.BasePacket) ([]byte, error) {
	config.LogDebug("Handle Capsule Message")
	tunnel, ok := h.peerRegistry.GetTunnelByPathID(packet.BaseHeader.ID)
	if !ok {
		return nil, fmt.Errorf("tunnel not found for path ID: %s", hex.EncodeToString(packet.BaseHeader.ID))
	}

	capsuleMessage, err := entity.ParseCapsuleMessage(packet, tunnel.EndKey())
	if err != nil {
		return nil, fmt.Errorf("failed to parse capsule message: %w", err)
	}

	return capsuleMessage.Payload, nil
}

func (h *handler) generateCapsuleMessage(data []byte, dst netip.Addr) ([]byte, *net.UDPAddr, error) {
	// TODO: AllowedIPsの仕組みを用いて実装する
	peer, ok := h.peerRegistry.GetPeerByVirtualAddr(dst)
	if !ok {
		return nil, nil, nil
	}

	tunnel, ok := peer.Tunnel()
	if !ok {
		return nil, nil, fmt.Errorf("failed to get tunnel from peer")
	}

	var packetType entity.TypeClass
	switch peer.NodeType() {
	case entity.NodeTypeInitiator, entity.NodeTypeLocalInitiator:
		packetType = entity.TypeClassCapsuleMessageFromResponder
	case entity.NodeTypeResponder, entity.NodeTypeLocalResponder:
		packetType = entity.TypeClassCapsuleMessageFromInitiator
	}

	capsuleMessage, err := entity.GenerateCapsuleMessage(tunnel.PathID(), peer.TransactionID(), data, packetType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate capsule message: %w", err)
	}

	buffer, err := capsuleMessage.Marshal(tunnel.EndKey())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal capsule message: %w", err)
	}

	// address determination process
	var sendAddress *net.UDPAddr
	if tunnel.ViaTRS() {
		switch h.localIPVersion {
		case entity.TypeLocalIPVersion4:
			sendAddress = tunnel.EndpointTRSv4()
		case entity.TypeLocalIPVersion6, entity.TypeLocalDualStackNetwork:
			sendAddress = tunnel.EndpointTRSv6()
		}
	} else {
		switch h.localIPVersion {
		case entity.TypeLocalIPVersion4:
			sendAddress = tunnel.Endpointv4()
		case entity.TypeLocalIPVersion6, entity.TypeLocalDualStackNetwork:
			sendAddress = tunnel.Endpointv6()
		}
	}

	config.LogDebug("Generated capsule message", "packet_type", packetType, "virtual_ip", dst, "real_ip", sendAddress)

	return buffer, sendAddress, nil
}
