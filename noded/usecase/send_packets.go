package usecase

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"golang.org/x/sync/errgroup"
)

func (h *handler) SendPackets(ctx context.Context,
	capsuleMessageQueue <-chan *entity.CapsuleMsgQueue,
	generatedPacketQueue <-chan entity.OutPacketQueue,
	directionRequestQueue <-chan entity.OutPacketQueue,
	reRegistrationQueue <-chan entity.OutPacketQueue,
	keepAliveQueue <-chan entity.OutPacketQueue,
) error {
	config.LogDebug("Starting send packets function")
	defer config.LogDebug("Finished send packets function")

	errGroup, ctx := errgroup.WithContext(ctx)

	err := h.socketStore.ClearDeadlines()
	if err != nil {
		return fmt.Errorf("failed to clear deadlines: %w", err)
	}

	stop := context.AfterFunc(ctx, func() {
		_ = h.socketStore.ExpireDeadlines()
	})
	defer stop()

	// TODO: 適切なSocketを選択する
	defaultSoc, err := h.socketStore.HighestPrioritySocket()
	if err != nil {
		return fmt.Errorf("failed to get highest priority socket: %w", err)
	}

	for i := 0; i < 1; i++ {
		errGroup.Go(func() error {
			for {
				var (
					queue entity.OutPacketQueue
				)

				select {
				case <-ctx.Done():
					return nil
				case q := <-capsuleMessageQueue:
					q.Lock() // FIXME: デッドロックする懸念
					if q.Buffer == nil || q.Addr == nil {
						continue
					}
					queue.Buffer = q.Buffer
					queue.Addr = q.Addr
					queue.Soc = q.Soc
				case queue = <-generatedPacketQueue:
				case queue = <-directionRequestQueue:
				case queue = <-reRegistrationQueue:
				case queue = <-keepAliveQueue:

				}

				if queue.Soc != nil {
					_, err = queue.Soc.WriteToUDP(queue.Buffer, queue.Addr)
					config.LogDebug("Sent packet", "src", queue.Soc.LocalAddr(), "dst", queue.Addr.String())
				} else {
					_, err = defaultSoc.WriteToUDP(queue.Buffer, queue.Addr)
					config.LogDebug("Sent packet", "src", defaultSoc.LocalAddr(), "dst", queue.Addr.String())
				}

				if err != nil {
					return fmt.Errorf("failed to write packet: %w", err)
				}
			}
		})
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("failed to send packets: %w", err)
	}

	return nil
}
