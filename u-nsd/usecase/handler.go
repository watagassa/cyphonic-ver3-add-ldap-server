//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/u-nsd/$GOPACKAGE
package usecase

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"

	"github.com/Pluslab/cyphonic/u-nsd/entity"
	"golang.org/x/sync/errgroup"

	"github.com/Pluslab/cyphonic/u-nsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/u-nsd/usecase/repository"
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

type Handler interface {
	Run(ctx context.Context) error
	ReceivePackets(ctx context.Context, receivePacketQueue chan PacketQueue, errGroup *errgroup.Group) error
	SendPackets(ctx context.Context, sendPacketQueue chan PacketQueue, errGroup *errgroup.Group) error
	HandlePackets(ctx context.Context, receivePacketQueue, sendPacketQueue chan PacketQueue, errGroup *errgroup.Group) error
}

type handler struct {
	udpConn    repository.UDPConn
	sqlHandler repository.SQLHandler
}

func NewHandler(udpConn repository.UDPConn, sqlHandler repository.SQLHandler) Handler {
	return &handler{
		udpConn:    udpConn,
		sqlHandler: sqlHandler,
	}
}

func (h *handler) Run(ctx context.Context) error {
	receivePacketQueue := make(chan PacketQueue, packetQueueSize)
	sendPacketQueue := make(chan PacketQueue, packetQueueSize)

	errGroup, ctx := errgroup.WithContext(ctx)

	if err := h.sqlHandler.CreateSelfNotificationService(); err != nil {
		return fmt.Errorf("failed to create self notification service: %w", err)
	}

	errGroup.Go(func() error {
		return h.ReceivePackets(ctx, receivePacketQueue, errGroup)
	})

	errGroup.Go(func() error {
		return h.SendPackets(ctx, sendPacketQueue, errGroup)
	})

	errGroup.Go(func() error {
		return h.HandlePackets(ctx, receivePacketQueue, sendPacketQueue, errGroup)
	})

	// Wait for context cancellation and gracefully shut down goroutines
	go func() {
		<-ctx.Done()
		h.shutdown()
	}()

	config.LogInfo("UDP Notification service is running...")

	return errGroup.Wait()
}

func (h *handler) shutdown() {
	if err := h.sqlHandler.DeleteNodeAddressesBySelfNSID(); err != nil {
		config.LogErr(fmt.Sprintf("failed to delete node addresses by self NSID: %v", err))
	}
	if err := h.sqlHandler.DeleteSelfNotificationService(); err != nil {
		config.LogErr(fmt.Sprintf("failed to delete self notification service: %v", err))
	}
	if err := h.sqlHandler.Close(); err != nil {
		config.LogErr(fmt.Sprintf("failed to close DB connection: %v", err))
	}
	if err := h.udpConn.Close(); err != nil {
		config.LogErr(fmt.Sprintf("failed to close UDP connection: %v", err))
	}
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

func (h *handler) parsePacket(buffer []byte) (*entity.BasePacket, error) {
	packet, err := entity.ParsePacket(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse packet: %w", err)
	}

	return packet, nil
}

func (h *handler) handle(parsedPacket *entity.BasePacket, addr *net.UDPAddr) ([]byte, *net.UDPAddr, error) {
	switch parsedPacket.BaseHeader.Type {
	case entity.TypeClassRegistrationRequest:
		return h.handleRegistrationRequest(parsedPacket, addr)
	case entity.TypeClassRouteDirectionToResponder:
		return h.handleRouteDirectionToResponder(parsedPacket)
	case entity.TypeClassDirectionRequest:
		return h.handleDirectionRequest(parsedPacket)
	case entity.TypeClassRouteDirectionConfirmation:
		return h.handleRouteDirectionConfirmation(parsedPacket)
	case entity.TypeClassRouteDirectionToInitiator:
		return h.handleRouteDirectionToInitiator(parsedPacket)
	case entity.TypeClassKeepAlive:
		return h.handleKeepAlive(parsedPacket, addr)
	default:
		return nil, nil, fmt.Errorf("%w: %d", ErrUnknownPacketType, parsedPacket.BaseHeader.Type)
	}
}

// handleRegistrationRequest processes a Registration Request packet from CYPHONIC Node
func (h *handler) handleRegistrationRequest(packet *entity.BasePacket, addr *net.UDPAddr) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("handle Registration Request")

	var initiatorNodeAddress *entity.NodeAddress

	// TODO: implement decryption
	registrationRequest, err := entity.ParseRegistrationRequest(packet, []byte{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse registration request: %w", err)
	}

	netipAddr, ok := netip.AddrFromSlice(addr.IP)
	if !ok {
		return nil, nil, fmt.Errorf("failed to convert net.IP to netip.Addr: %w", err)
	}

	if netipAddr.Is4In6() || netipAddr.Is4() {
		// addr is IPv4
		// FIXME: interfaceID is 0
		initiatorNodeAddress, err = h.sqlHandler.SaveInitiatorNodeAddress("0", registrationRequest, addr.IP.To4(), net.IPv6zero, addr.Port)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to save initiator node address: %w", err)
		}
	} else if netipAddr.Is6() {
		// addr is IPv6
		// FIXME: interfaceID is 1
		initiatorNodeAddress, err = h.sqlHandler.SaveInitiatorNodeAddress("1", registrationRequest, net.IPv4zero, addr.IP.To16(), addr.Port)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to save initiator node address: %w", err)
		}
	}

	registrationResponse, err := entity.GenerateRegisrationResponse(registrationRequest, []byte{}, initiatorNodeAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate registration response: %w", err)
	}

	buf, err := registrationResponse.ConvertBytes()
	if err != nil {
		config.LogErr("Failed to convert to bytes", "error", err)
	}

	return buf, addr, nil
}

// handleRouteDirectionToResponder processes a Route Direction packet from DS to Responder Node
func (h *handler) handleRouteDirectionToResponder(packet *entity.BasePacket) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("handle Route Direction To Responder")

	err := packet.ChangeNoHMAC()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to change no HMAC: %w", err)
	}

	routeDirection, err := entity.ParseRouteDirection(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse route direction to responder: %w", err)
	}

	// TODO: implement encryption
	buffer, err := routeDirection.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal Route Direction Confirmation: %w", err)
	}

	var addr *net.UDPAddr
	if !net.IP(routeDirection.NATResponderIPv4[:]).IsUnspecified() {
		addr = &net.UDPAddr{
			IP:   routeDirection.NATResponderIPv4[:],
			Port: int(routeDirection.NATResponderPort),
		}
	} else {
		addr = &net.UDPAddr{
			IP:   routeDirection.NATResponderIPv6[:],
			Port: int(routeDirection.NATResponderPort),
		}
	}

	config.LogDebug("Sent Route Direction To Responder", "dst", net.JoinHostPort(addr.IP.String(), fmt.Sprintf("%d", addr.Port)))

	return buffer, addr, nil
}

// handleRouteDirectionToInitiator processes a Route Direction packet from DS to Initiator Node
func (h *handler) handleRouteDirectionToInitiator(packet *entity.BasePacket) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("handle Route Direction To Initiator")

	err := packet.ChangeNoHMAC()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to change no HMAC: %w", err)
	}

	routeDirection, err := entity.ParseRouteDirection(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse route direction to initiator: %w", err)
	}

	// TODO: implement encryption
	buffer, err := routeDirection.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal Route Direction Confirmation: %w", err)
	}

	var addr *net.UDPAddr
	if !net.IP(routeDirection.NATInitiatorIPv4[:]).IsUnspecified() {
		addr = &net.UDPAddr{
			IP:   routeDirection.NATInitiatorIPv4[:],
			Port: int(routeDirection.NATInitiatorPort),
		}
	} else {
		addr = &net.UDPAddr{
			IP:   routeDirection.NATInitiatorIPv6[:],
			Port: int(routeDirection.NATInitiatorPort),
		}
	}

	config.LogDebug("Sent Route Direction To Initiator", "dst", net.JoinHostPort(addr.IP.String(), fmt.Sprintf("%d", addr.Port)))

	return buffer, addr, nil
}

// handleDirectionRequest processes a Direction Request packet from Initiator Node to DS
func (h *handler) handleDirectionRequest(packet *entity.BasePacket) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Received Direction Request")

	// TODO: implement decryption
	prevDirectionRequest, err := entity.ParseDirectionRequest(packet, []byte{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse registration request: %w", err)
	}

	directionRequest := entity.GenerateDirectionRequest(prevDirectionRequest)

	buf, err := directionRequest.ConvertBytes()
	if err != nil {
		config.LogErr("Failed to convert to bytes", "error", err)
	}

	return buf, h.udpConn.GetDSAddr(), nil
}

// handleRouteDirectionConfirmation processes a Route Direction Confirmation packet from Responder Node to DS
func (h *handler) handleRouteDirectionConfirmation(packet *entity.BasePacket) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Received Route Direction Confirmation")

	// TODO: implement decryption
	routeDirection, err := entity.ParseRouteDirection(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse route direction confirmation: %w", err)
	}

	buffer, err := routeDirection.MarshalWithoutHMAC()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal Route Direction Confirmation: %w", err)
	}

	return buffer, h.udpConn.GetDSAddr(), nil
}

// handleKeepAlive sends ACK for Keep Alive packet from CYPHONIC Node
func (h *handler) handleKeepAlive(packet *entity.BasePacket, addr *net.UDPAddr) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Handling Keep Alive")

	// TODO: Decode HMAC

	// Generate ACK packet
	ack := entity.GenerateKeepAliveACK(packet.BaseHeader.ID, packet.BaseHeader)

	// Convert ACK packet to bytes
	buf, err := ack.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert ACK to bytes: %w", err)
	}

	return buf, addr, nil
}
