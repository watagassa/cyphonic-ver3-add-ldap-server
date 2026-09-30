package usecase

import (
	"context"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
)

func (h *handler) StageOutboundMessages(ctx context.Context,
	outboundMsgQueue <-chan entity.OutboundMsgQueue,
	stagedOutboundMsgQueue chan<- entity.StagedOutboundMsgQueue,
	capsuleMessageQueue chan<- *entity.CapsuleMsgQueue,
) error {
	config.LogDebug("Starting stage outbound messages function")
	defer config.LogDebug("Finished stage outbound messages function")

	for {
		var msg entity.OutboundMsgQueue

		select {
		case <-ctx.Done():
			return nil
		case msg = <-outboundMsgQueue:
		}

		if msg.Data == nil || msg.Dst == nil {
			continue
		}

		emptyQueue := entity.CapsuleMsgQueue{}
		emptyQueue.Lock()

		msgQueue := entity.StagedOutboundMsgQueue{
			Data:     msg.Data,
			Dst:      msg.Dst,
			QueuePtr: &emptyQueue,
		}

		// TODO: expired処理の実装
		// pre-send the empty queue
		capsuleMessageQueue <- &emptyQueue
		stagedOutboundMsgQueue <- msgQueue
	}
}

func (h *handler) StageInboundMessages(ctx context.Context,
	receivedPacketQueue <-chan entity.InPacketQueue,
	parsedPacketQueue chan<- entity.ParsedPacketQueue,
	stagedInboundMsgQueue chan<- entity.StagedInboundMsgQueue,
	inboundMsgQueue chan<- *entity.InboundMsgQueue,
) error {
	config.LogDebug("Starting stage inbound messages function")
	defer config.LogDebug("Finished stage inbound messages function")

	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-receivedPacketQueue:
			parsedPacket, err := h.parsePacket(packet.Buffer)
			if err != nil {
				config.LogErr("Failed to parse packet", "error", err)
				continue
			}

			if parsedPacket.BaseHeader.Type != entity.TypeClassCapsuleMessageFromInitiator &&
				parsedPacket.BaseHeader.Type != entity.TypeClassCapsuleMessageFromResponder {
				parsedPacketQueue <- entity.ParsedPacketQueue{
					Packet: parsedPacket,
					Addr:   packet.Addr,
				}
				continue
			}

			emptyQueue := entity.InboundMsgQueue{}
			emptyQueue.Lock()

			msgQueue := entity.StagedInboundMsgQueue{
				Packet:   parsedPacket,
				QueuePtr: &emptyQueue,
			}

			// TODO: expired処理の実装
			// pre-send the empty queue
			inboundMsgQueue <- &emptyQueue
			stagedInboundMsgQueue <- msgQueue
		}
	}
}
