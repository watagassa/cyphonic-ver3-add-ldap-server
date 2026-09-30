//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/dsd/$GOPACKAGE
package usecase

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/dsd/entity"
	"github.com/Pluslab/cyphonic/dsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/dsd/usecase/repository"
	"golang.org/x/sync/errgroup"
)

const (
	packetQueueSize = 65535
	maxPacketSize   = 4096
	numWorkers      = 4
)

var ErrUnknownPacketType = errors.New("unknown packet type")

type PacketQueue struct {
	Buffer []byte
	Addr   *net.UDPAddr
}

var _ Handler = (*handler)(nil)

type Handler interface {
	Run(ctx context.Context) error
	ReceivePackets(ctx context.Context, receivePacketQueue chan PacketQueue, errGroup *errgroup.Group) error
	SendPackets(ctx context.Context, sendPacketQueue chan PacketQueue, errGroup *errgroup.Group) error
	HandlePackets(ctx context.Context, receivePacketQueue, sendPacketQueue chan PacketQueue, errGroup *errgroup.Group) error
}

type handler struct {
	udpConn    repository.UDPConn
	sqlHandler repository.SQLHandler
	// redisHandler repository.RedisClient
}

func NewHandler(udpConn repository.UDPConn, sqlHandler repository.SQLHandler) (Handler, error) {
	return &handler{
		udpConn:    udpConn,
		sqlHandler: sqlHandler,
	}, nil
}

func (h *handler) Run(ctx context.Context) error {
	receivePacketQueue := make(chan PacketQueue, packetQueueSize)
	sendPacketQueue := make(chan PacketQueue, packetQueueSize)

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return h.ReceivePackets(ctx, receivePacketQueue, errGroup)
	})

	errGroup.Go(func() error {
		return h.SendPackets(ctx, sendPacketQueue, errGroup)
	})

	errGroup.Go(func() error {
		return h.HandlePackets(ctx, receivePacketQueue, sendPacketQueue, errGroup)
	})

	// Gracefully shutdown
	go func() {
		<-ctx.Done()

		config.LogDebug("Closing UDP connection")
		if err := h.udpConn.Close(); err != nil {
			config.LogErr("Failed to close UDP connection", "error", err)
		}

		config.LogDebug("Closing DB connection")
		if err := h.sqlHandler.Close(); err != nil {
			config.LogErr("Failed to close DB connection", "error", err)
		}
	}()

	config.LogInfo("Direction service is running...")

	return errGroup.Wait()
}

func (h *handler) ReceivePackets(ctx context.Context, receivePacketQueue chan PacketQueue, errGroup *errgroup.Group) error {
	for i := 0; i < numWorkers; i++ {
		errGroup.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					buffer := make([]byte, maxPacketSize)
					packetLength, address, err := h.udpConn.ReadFromUDP(buffer)
					if err != nil {
						if errors.Is(err, net.ErrClosed) {
							return nil
						}
						return fmt.Errorf("failed to read from UDP: %w", err)
					}

					if packetLength > maxPacketSize {
						return fmt.Errorf("packet length exceeds maximum size: %d", packetLength)
					}

					packet := PacketQueue{
						Buffer: buffer[:packetLength],
						Addr:   address,
					}
					receivePacketQueue <- packet
				}
			}
		})
	}
	return nil
}

func (h *handler) SendPackets(ctx context.Context, sendPacketQueue chan PacketQueue, errGroup *errgroup.Group) error {
	for i := 0; i < numWorkers; i++ {
		errGroup.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case packet := <-sendPacketQueue:
					_, err := h.udpConn.WriteToUDP(packet.Buffer, packet.Addr)
					if err != nil {
						return fmt.Errorf("failed to write packet: %w", err)
					}
				}
			}
		})
	}
	return nil
}

func (h *handler) HandlePackets(ctx context.Context, receivePacketQueue, sendPacketQueue chan PacketQueue, errGroup *errgroup.Group) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-receivePacketQueue:
			parsedPacket, err := h.parsePacket(packet.Buffer)
			if err != nil {
				config.LogErr("Failed to parse packet", "error", err)
				continue
			}

			buf, sendAddress, err := h.handle(parsedPacket, packet.Addr)
			if err != nil {
				config.LogErr("Failed to handle packet", "error", err)
				continue
			}

			if buf != nil && sendAddress != nil {
				sendPacketQueue <- PacketQueue{
					Buffer: buf,
					Addr:   sendAddress,
				}
			}
		}
	}
}

func (h *handler) parsePacket(buffer []byte) (*entity.BasePacket, error) {
	basePacket, err := entity.ParsePlainPacket(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plain packet: %w", err)
	}

	return basePacket, nil
}

func (h *handler) handle(packet *entity.BasePacket, addr *net.UDPAddr) ([]byte, *net.UDPAddr, error) {
	switch packet.BaseHeader.Type {
	case entity.TypeClassDirectionRequest:
		return h.handleDirectionRequest(packet)
	case entity.TypeClassRouteDirectionConfirmation:
		return h.handleRouteDirectionConfirmation(packet, addr)
	default:
		return nil, nil, fmt.Errorf("unknown packet type: %d", packet.BaseHeader.Type)
	}
}

func (h *handler) handleDirectionRequest(packet *entity.BasePacket) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("handle Direction Request")

	directionRequest, err := entity.ParseDirectionRequest(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse direction request: %w", err)
	}

	initiatorNodeInformation, initiatorNodeAddress, err := h.sqlHandler.GetNodeInformationAndNodeAddressByNodeID(fmt.Sprintf("\\x%x", packet.BaseHeader.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get node information and node address by node ID: %w", err)
	}

	responderNodeInformation, responderNodeAddress, err := h.sqlHandler.GetNodeInformationAndNodeAddressByFQDN(string(directionRequest.FQDN))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get node information and node address by FQDN: %w", err)
	}

	responderNotificationService, err := h.sqlHandler.GetNotificationServiceByNSID(responderNodeAddress.NSID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get notification service by NSID: %w", err)
	}

	tunnelKey, err := entity.GenerateCommonKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tunnel key: %w", err)
	}

	var processCode entity.ProcessCodeClass
	trs := entity.TunnelRelayService{
		TunnleRelayServiceIPv4: net.IPv4zero,
		TunnleRelayServiceIPv6: net.IPv6zero,
		TunnleRelayServicePort: 0,
	}

	// Check if TRS is needed
	if initiatorNodeAddress.ShouldRelay(&responderNodeAddress) {
		processCode = entity.TunnelRequestToTRS
	} else {
		// Check if UDP Hole Punching is needed (Check if Responder side NAPT exists)
		if initiatorNodeAddress.ShouldHolePunching(&responderNodeAddress) {
			processCode = entity.TunnelRequestToNATResponder
		} else {
			processCode = entity.TunnelRequestToResponder
		}
	}

	// processCode = entity.TunnelRequestToTRS // Comment out to force TRS for testing

	// Execute the required operations based on the processcode.
	switch processCode {
	// If TRS is needed, get random TRS from DB and save path information to DB
	case entity.TunnelRequestToTRS:
		// Get random TRS from DB
		trs, err = h.sqlHandler.GetRandomUDPTunnelRelayService()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get random UDP tunnel relay service: %w", err)
		}

		// Save path information to DB
		// FIXME: radis.go未使用のため仮実装
		err = h.sqlHandler.SetPathInformation(directionRequest.PathID, tunnelKey, initiatorNodeAddress, responderNodeAddress)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to set path information: %w", err)
		}
	}

	// Generate Route Direction
	routeDirection, err := entity.GenerateRouteDirection(
		&packet.BaseHeader,
		&directionRequest.PathID,
		processCode,
		tunnelKey,
		&initiatorNodeInformation,
		&responderNodeInformation,
		&initiatorNodeAddress,
		&responderNodeAddress,
		&entity.TunnelRelayService{
			// TRS Address or zero value if not needed
			TunnleRelayServiceIPv4: trs.TunnleRelayServiceIPv4,
			TunnleRelayServiceIPv6: trs.TunnleRelayServiceIPv6,
			TunnleRelayServicePort: trs.TunnleRelayServicePort,
		},
		string(directionRequest.FQDN),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate route direction: %w", err)
	}

	buffer, err := routeDirection.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal route direction: %w", err)
	}

	// Default to IPv4 address
	sendAddress := &net.UDPAddr{
		IP:   responderNotificationService.NSIPv4,
		Port: responderNotificationService.NSPort,
	}

	// If IPv6 address is specified, use IPv6 address
	if !net.IPv6zero.Equal(responderNotificationService.NSIPv6) {
		sendAddress.IP = responderNotificationService.NSIPv6
		config.LogDebug("Using IPv6 address for notification service", "address", sendAddress.IP.String())
	}

	return buffer, sendAddress, nil
}

func (h *handler) handleRouteDirectionConfirmation(packet *entity.BasePacket, addr *net.UDPAddr) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("handle Route Direction Confirmation")
	routeDirectionConfirmation, err := entity.ParseRouteDirection(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse route direction confirmation: %w", err)
	}

	routeDirectionConfirmation.ChangeRouteDirectionType(entity.TypeClassRouteDirectionToInitiator)

	buffer, err := routeDirectionConfirmation.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal route direction confirmation: %w", err)
	}

	// Get Initiator Info
	_, initiatorNodeAddress, err := h.sqlHandler.GetNodeInformationAndNodeAddressByFQDN(string(routeDirectionConfirmation.InitiatorFQDN))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get node information and node address by node ID: %w", err)
	}

	// Get Initiator Side NS
	initiatorNotificationService, err := h.sqlHandler.GetNotificationServiceByNSID(initiatorNodeAddress.NSID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get notification service by NSID: %w", err)
	}

	// Default to IPv4 address
	sendAddress := &net.UDPAddr{
		IP:   initiatorNotificationService.NSIPv4,
		Port: initiatorNotificationService.NSPort,
	}

	// If IPv6 address is specified, use IPv6 address
	if !net.IPv6zero.Equal(initiatorNotificationService.NSIPv6) {
		sendAddress.IP = initiatorNotificationService.NSIPv6
		config.LogDebug("Using IPv6 address for notification service", "address", sendAddress.IP.String())
	}

	return buffer, sendAddress, nil
}
