package usecase

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"

	"github.com/Pluslab/cyphonic/u-trsd/entity"
	"github.com/Pluslab/cyphonic/u-trsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/u-trsd/usecase/repository"
	"golang.org/x/sync/errgroup"
)

const (
	packetQueueSize = 65535
	maxPacketSize   = 4096
)

// numWorkers is set to the number of CPU cores available for parallel processing.
var numWorkers = runtime.NumCPU()

// ErrUnknownPacketType is an error that indicates that the packet type is unknown
var ErrUnknownPacketType = errors.New("unknown packet type")

// PacketQueue is a struct that holds the packet data and the address to send it to.
type PacketQueue struct {
	Buffer []byte
	Addr   *net.UDPAddr
}

var _ Handler = (*handler)(nil)

// Handler is an interface that defines the methods for handling packets.
type Handler interface {
	Run(ctx context.Context) error
}

type handler struct {
	udpConn    repository.UDPConn
	sqlHandler repository.SQLHandler
	pathCache  repository.PathCache
}

// NewHandler creates a new handler with the provided UDP connection, SQL handler, and path cache.
func NewHandler(udpConn repository.UDPConn, sqlHandler repository.SQLHandler, pathCache repository.PathCache) Handler {
	return &handler{
		udpConn:    udpConn,
		sqlHandler: sqlHandler,
		pathCache:  pathCache,
	}
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

	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Instance registration if enabled in config
	if cfg.InstanceRegistration {
		err = h.registerInstance()
		if err != nil {
			return fmt.Errorf("failed to register instance: %w", err)
		}
		config.LogInfo("Instance registered successfully", "FQDN", cfg.FQDN)
	}

	// Gracefully shutdown
	go func() {
		<-ctx.Done()

		// Instance deregistration if enabled in config
		if cfg.InstanceRegistration {
			config.LogDebug("Deregistering instance.", "FQDN", cfg.FQDN)
			err := h.deregisterInstance()
			if err != nil {
				config.LogErr("Failed to deregister instance", "error", err)
			} else {
				config.LogInfo("Instance deregistered successfully", "FQDN", cfg.FQDN)
			}
		}

		config.LogDebug("Closing UDP connection.")
		if err := h.udpConn.Close(); err != nil {
			config.LogErr("Failed to close UDP connection", "error", err)
		}

		config.LogDebug("Closing DB connection")
		if err := h.sqlHandler.Close(); err != nil {
			config.LogErr("Failed to close DB connection", "error", err)
		}
	}()

	config.LogInfo("UDP Tunnel Relay service is running...")

	return errGroup.Wait()
}

func (h *handler) registerInstance() error {
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to convert port to integer: %v", err)
	}

	trs := entity.TunnelRelayService{
		TunnleRelayServiceID:   cfg.FQDN,
		TunnleRelayServiceIPv4: net.ParseIP(cfg.TRSIPv4).To4(),
		TunnleRelayServiceIPv6: net.ParseIP(cfg.TRSIPv6).To16(),
		TunnleRelayServicePort: port,
		QUICFlag:               false,
	}

	err = h.sqlHandler.CreateTunnelRelayService(trs)
	if err != nil {
		return fmt.Errorf("failed to instance registration: %w", err)
	}

	return nil
}

func (h *handler) deregisterInstance() error {
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	err = h.sqlHandler.DeleteTunnelRelayService(cfg.FQDN)
	if err != nil {
		return fmt.Errorf("failed to instance deregistration: %w", err)
	}

	return nil
}

func (h *handler) ReceivePackets(ctx context.Context, receivePacketQueue chan PacketQueue, errGroup *errgroup.Group) error {
	config.LogDebug("Starting packet receiver", "numWorkers", 1)
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
	return nil
}

func (h *handler) SendPackets(ctx context.Context, sendPacketQueue chan PacketQueue, errGroup *errgroup.Group) error {
	config.LogDebug("Starting packet sender workers", "numWorkers", 1)
	errGroup.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case packet := <-sendPacketQueue:
				_, err := h.udpConn.WriteToUDP(packet.Buffer, packet.Addr)
				if err != nil {
					config.LogErr("Failed to write packet", "error", err, "buffer", packet.Buffer, "to", packet.Addr)
					return fmt.Errorf("failed to write packet: %w", err)
				}
				config.LogDebug("Sent packet", "buffer", packet.Buffer, "to", packet.Addr)
			}
		}
	})
	return nil
}

func (h *handler) HandlePackets(ctx context.Context, receivePacketQueue, sendPacketQueue chan PacketQueue, errGroup *errgroup.Group) error {
	config.LogDebug("Starting packet handler workers", "numWorkers", numWorkers)
	for i := 0; i < numWorkers; i++ {
		errGroup.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case packet := <-receivePacketQueue:
					config.LogDebug("Packet Received", "buffer", packet.Buffer, "from", packet.Addr, "goroutineCount", runtime.NumGoroutine())
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
		})
	}
	return nil
}

func (h *handler) parsePacket(buffer []byte) (*entity.BasePacket, error) {
	basePacket, err := entity.ParsePlainPacket(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plain packet: %w", err)
	}

	return basePacket, nil
}

// return: Buffer and Address to send the response packet.
func (h *handler) handle(packet *entity.BasePacket, srcAddr *net.UDPAddr) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Handling packet", "type", packet.BaseHeader.Type, "srcAddr", srcAddr)
	switch packet.BaseHeader.Type {
	// Route Selection Process
	case entity.TypeClassHolePunching:
		return h.handleHolePunching(packet, srcAddr)

	// Tunnel Construction Process
	case entity.TypeClassTunnelRequestForUDP:
		return h.handleTunnelRequest(packet, srcAddr)
	case entity.TypeClassTunnelResponseForUDP:
		return h.handleTunnelResponse(packet)

	// Tunnel Communication
	case entity.TypeClassCapsuleMessageFromInitiator:
		return h.handleCapsuleMessageFromInitiator(packet)
	case entity.TypeClassCapsuleMessageFromResponder:
		return h.handleCapsuleMessageFromResponder(packet)

	case entity.TypeClassKeepAlive:
		config.LogDebug("KeepAlive packet received, no action taken", "packet", packet)
		return nil, nil, nil

	default:
		return nil, nil, fmt.Errorf("unknown packet type: %d", packet.BaseHeader.Type)
	}
}

func (h *handler) handleHolePunching(packet *entity.BasePacket, srcAddr *net.UDPAddr) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Handling hole punching packet", "packet", packet)

	// Get HMAC	field
	hmac, err := entity.GetHMACField(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get HMAC field: %w", err)
	}

	// Decode the HMAC
	hmacFlg, err := entity.DecodeHMAC(hmac, packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode HMAC: %w", err)
	}
	if !hmacFlg {
		return nil, nil, fmt.Errorf("HMAC verification failed")
	}

	// Set the responder node address in the path cache
	pathID := string(packet.BaseHeader.ID)
	if err := h.pathCache.UpdateResponderNodeAddr(pathID, srcAddr.IP, uint16(srcAddr.Port)); err != nil {
		return nil, nil, fmt.Errorf("failed to update responder node address in internal cache: %w", err)
	}

	// Set the responder node address in the database
	if err := h.sqlHandler.SetResponderNodeAddress(pathID, srcAddr.IP, uint16(srcAddr.Port)); err != nil {
		return nil, nil, fmt.Errorf("failed to set responder node address in db: %w", err)
	}

	return nil, nil, nil
}

func (h *handler) handleTunnelRequest(packet *entity.BasePacket, srcAddr *net.UDPAddr) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Handling tunnel request packet", "packet", packet)

	// Get HMAC	field
	hmac, err := entity.GetHMACField(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get HMAC field: %w", err)
	}

	// Decode the HMAC
	hmacFlg, err := entity.DecodeHMAC(hmac, packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode HMAC: %w", err)
	}
	if !hmacFlg {
		return nil, nil, fmt.Errorf("HMAC verification failed")
	}

	// Set the initiator node address in the path cache
	pathID := string(packet.BaseHeader.ID)
	if err := h.pathCache.UpdateInitiatorNodeAddr(pathID, srcAddr.IP, uint16(srcAddr.Port)); err != nil {
		return nil, nil, fmt.Errorf("failed to update initiator node address in internal cache: %w", err)
	}

	// Set the initiator node address in the database
	if err := h.sqlHandler.SetInitiatorNodeAddress(pathID, srcAddr.IP, uint16(srcAddr.Port)); err != nil {
		return nil, nil, fmt.Errorf("failed to set initiator node address in db: %w", err)
	}

	// Send Tunnel Request to the responder
	buf, err := packet.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal Tunnel Request: %w", err)
	}

	// Get the responder node address from the path cache
	_, rnAddr, rnPort, err := h.pathCache.GetResponderNodeAddr(pathID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get responder node address: %w", err)
	}

	sendAddress := &net.UDPAddr{
		IP:   rnAddr,
		Port: int(rnPort),
	}

	return buf, sendAddress, nil
}

func (h *handler) handleTunnelResponse(packet *entity.BasePacket) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Handling tunnel response packet", "packet", packet)

	// Get HMAC	field
	hmac, err := entity.GetHMACField(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get HMAC field: %w", err)
	}

	// Decode the HMAC
	hmacFlg, err := entity.DecodeHMAC(hmac, packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode HMAC: %w", err)
	}
	if !hmacFlg {
		return nil, nil, fmt.Errorf("HMAC verification failed")
	}

	// Send Tunnel Response to the initiator
	buf, err := packet.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal Tunnel Request: %w", err)
	}

	// Get the responder node address from the path cache
	pathID := string(packet.BaseHeader.ID)
	_, inAddr, inPort, err := h.pathCache.GetInitiatorNodeAddr(pathID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get responder node address: %w", err)
	}

	sendAddress := &net.UDPAddr{
		IP:   inAddr,
		Port: int(inPort),
	}

	return buf, sendAddress, nil
}

func (h *handler) handleCapsuleMessageFromInitiator(packet *entity.BasePacket) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Handling capsule message from initiator packet", "packet", packet)

	// Get HMAC	field
	hmac, err := entity.GetHMACField(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get HMAC field: %w", err)
	}

	// Decode the HMAC
	hmacFlg, err := entity.DecodeHMAC(hmac, packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode HMAC: %w", err)
	}
	if !hmacFlg {
		return nil, nil, fmt.Errorf("HMAC verification failed")
	}

	// Get responder node address from path cache
	_, rnAddr, rnPort, err := h.pathCache.GetResponderNodeAddr(string(packet.BaseHeader.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get responder node address: %w", err)
	}

	// Set the responder node to the packet from the initiator
	sendAddress := &net.UDPAddr{
		IP:   rnAddr,
		Port: int(rnPort),
	}

	sendPacket, err := packet.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal capsule message packet: %w", err)
	}

	// Update the path cache expiration time
	h.pathCache.Update(string(packet.BaseHeader.ID))

	return sendPacket, sendAddress, nil
}

func (h *handler) handleCapsuleMessageFromResponder(packet *entity.BasePacket) ([]byte, *net.UDPAddr, error) {
	config.LogDebug("Handling capsule message from responder packet", "packet", packet)

	// Get HMAC	field
	hmac, err := entity.GetHMACField(packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get HMAC field: %w", err)
	}

	// Decode the HMAC
	hmacFlg, err := entity.DecodeHMAC(hmac, packet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode HMAC: %w", err)
	}
	if !hmacFlg {
		return nil, nil, fmt.Errorf("HMAC verification failed")
	}

	// Get the initiator node address from the path cache
	_, inAddr, inPort, err := h.pathCache.GetInitiatorNodeAddr(string(packet.BaseHeader.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get responder node address: %w", err)
	}

	// Set the responder node to the packet from the initiator
	sendAddress := &net.UDPAddr{
		IP:   inAddr,
		Port: int(inPort),
	}

	sendPacket, err := packet.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal capsule message packet: %w", err)
	}

	// Update the path cache expiration time
	h.pathCache.Update(string(packet.BaseHeader.ID))

	return sendPacket, sendAddress, nil
}
