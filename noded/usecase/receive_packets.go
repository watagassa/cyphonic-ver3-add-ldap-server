package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"golang.org/x/sync/errgroup"
)

func (h *handler) ReceivePackets(ctx context.Context, receivedPacketQueue chan<- entity.InPacketQueue) error {
	config.LogDebug("Starting receive packets function")
	defer config.LogDebug("Finished receive packets function")

	errGroup, ctx := errgroup.WithContext(ctx)

	err := h.socketStore.ClearDeadlines()
	if err != nil {
		return fmt.Errorf("failed to clear deadlines: %w", err)
	}

	stop := context.AfterFunc(ctx, func() {
		_ = h.socketStore.ExpireDeadlines()
	})
	defer stop()

	for _, soc := range h.socketStore.Values() {
		// a for loop to change the number of threads
		// TODO: 可能なら並列化(読み取った順にTUNに書き込む縛りがあるため多分無理)
		for i := 0; i < 1; i++ {
			errGroup.Go(func() error {
				for {
					buffer := make([]byte, maxPacketSize)
					packetLength, address, err := soc.ReadFromUDP(buffer)

					if err != nil {
						if errors.Is(err, os.ErrDeadlineExceeded) {
							config.LogDebug("Read deadline exceeded, stopping packet reception")
							return nil
						}
						return fmt.Errorf("failed to read from UDP: %w", err)
					}

					if packetLength > maxPacketSize {
						return fmt.Errorf("packet length exceeds maximum size: %d", packetLength)
					}

					packet := entity.InPacketQueue{
						Buffer: buffer[:packetLength],
						Addr:   address,
					}

					select {
					case <-ctx.Done():
						return nil
					case receivedPacketQueue <- packet:
					}
				}
			})
		}
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("failed to receive packets: %w", err)
	}

	return nil
}
