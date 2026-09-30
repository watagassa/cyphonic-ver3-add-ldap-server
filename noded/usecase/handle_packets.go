package usecase

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
)

func (h *handler) HandlePackets(ctx context.Context,
	parsedPacketQueue <-chan entity.ParsedPacketQueue,
	generatedPacketQueue chan<- entity.OutPacketQueue,
	dnsAnswerQueue chan<- entity.DNSAnswerQueue,
	keepAliveAckQueue chan<- struct{},
) error {
	config.LogDebug("Starting handle packets function")
	defer config.LogDebug("Finished handle packets function")

	for {
		var (
			queues []*entity.OutPacketQueue
			err    error
		)
		select {
		case <-ctx.Done():
			return nil
		case q := <-parsedPacketQueue:
			switch q.Packet.BaseHeader.Type {
			case entity.TypeClassTunnelResponseForUDP:
				var answer *entity.DNSAnswerQueue
				answer, queues, err = h.handleTunnelResponseForUDP(q.Packet, q.Addr)
				if err != nil {
					config.LogErr("Failed to handle Tunnel Response For UDP", "error", err)
					continue
				}
				if answer != nil {
					dnsAnswerQueue <- *answer
				}

			case entity.TypeClassKeepAliveAck:
				if h.isKeepAliveAckFromNS(q.Addr) {
					select {
					case keepAliveAckQueue <- struct{}{}:
					default:
					}
				}

			default:
				queues, err = h.handle(q.Packet, q.Addr)
				if err != nil {
					config.LogErr("Failed to handle packet", "error", err)
				}
			}

			for _, queue := range queues {
				if queue.Buffer != nil && queue.Addr != nil {
					generatedPacketQueue <- *queue
				}
			}
		}
	}
}

func (h *handler) parsePacket(buffer []byte) (*entity.BasePacket, error) {
	packet, err := entity.ParsePacket(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse packet: %w", err)
	}

	return packet, nil
}

func (h *handler) handle(packet *entity.BasePacket, addr *net.UDPAddr) ([]*entity.OutPacketQueue, error) {
	switch packet.BaseHeader.Type {
	case entity.TypeClassRegistrationResponse:
		config.LogDebug("Handle Registration Response")
	case entity.TypeClassRouteDirectionToResponder:
		return h.handleRouteDirectionToResponder(packet)
	case entity.TypeClassRouteDirectionToInitiator:
		return h.handleRouteDirectionToInitiator(packet)
	case entity.TypeClassTunnelRequestForUDP:
		return h.handleTunnelRequestForUDP(packet, addr)
	case entity.TypeClassHolePunching:
		return h.handleHolePunching(packet, addr)
	case entity.TypeClassHolePunchingForOptimization:
		return h.handleHolePunchingForOptimization(packet, addr)
	case entity.TypeClassAckForOptimization:
		return h.handleAckForOptimization(packet, addr)
	case entity.TypeClassKeepAlive:
		break

	default:
		return nil, fmt.Errorf("%w: %d", ErrUnknownPacketType, packet.BaseHeader.Type)
	}

	return nil, nil
}

func (h *handler) handleRouteDirectionToResponder(packet *entity.BasePacket) ([]*entity.OutPacketQueue, error) {
	config.LogDebug("Handle Route Direction To Responder")

	routeDirection, err := entity.ParseRouteDirection(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse route direction to responder: %w", err)
	}

	if !bytes.Equal(routeDirection.BaseHeader.ID, h.nodeID) {
		return nil, fmt.Errorf("detected invalid NodeID: %x", routeDirection.BaseHeader.ID)
	}

	err = h.routeDirectionCache.Put(
		hex.EncodeToString(routeDirection.PathID),
		routeDirection,
		time.Now().Add(60*time.Second).UnixNano())

	if err != nil {
		return nil, fmt.Errorf("failed to put route direction to cache: %w", err)
	}

	routeDirection.ChangeRouteDirectionType(entity.TypeClassRouteDirectionConfirmation, h.nodeID)

	buffer, err := routeDirection.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Route Direction Confirmation: %w", err)
	}

	sendAddress, err := h.nsAddress.UDPAddr(h.localIPVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get NS UDP address: %w", err)
	}
	var generatedPackets []*entity.OutPacketQueue

	generatedPackets = append(generatedPackets,
		&entity.OutPacketQueue{
			Buffer: buffer,
			Addr:   &sendAddress,
		})

	config.LogDebug("Generated Route Direction Confirmation",
		"transaction_id", routeDirection.BaseHeader.TransactionID,
		"path_id", routeDirection.BaseHeader.ID,
		"route_direction_confirmation", routeDirection)

	// no hole punching
	if routeDirection.ProcessCode == entity.TunnelRequestToResponder {
		return generatedPackets, nil
	}

	// with hole punching
	holePunching := entity.GenerateHolePunching(routeDirection.PathID, h.baseHeader)
	buffer, err = holePunching.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Hole Punching: %w", err)
	}

	var natInitiatorAddress *net.UDPAddr
	switch h.localIPVersion {
	case entity.TypeLocalIPVersion4:
		if net.IP(routeDirection.NATInitiatorIPv4[:]).IsUnspecified() {
			config.LogDebug("Skip sending Hole Punching to NAT Initiator because of different IP version")
			break
		}
		// to NAT Initiator
		natInitiatorAddress = &net.UDPAddr{
			IP:   routeDirection.NATInitiatorIPv4[:],
			Port: int(routeDirection.NATInitiatorPort),
		}

	case entity.TypeLocalIPVersion6, entity.TypeLocalDualStackNetwork:
		if net.IP(routeDirection.NATInitiatorIPv6[:]).IsUnspecified() {
			config.LogDebug("Skip sending Hole Punching to NAT Initiator because of different IP version")
			break
		}
		// to NAT Initiator
		natInitiatorAddress = &net.UDPAddr{
			IP:   routeDirection.NATInitiatorIPv6[:],
			Port: int(routeDirection.NATInitiatorPort),
		}
	}

	// specify the destination for hole punching
	if natInitiatorAddress != nil {
		// to NAT Initiator
		generatedPackets = append(generatedPackets,
			&entity.OutPacketQueue{
				Buffer: buffer,
				Addr:   natInitiatorAddress,
			})
	}
	config.LogDebug("Generated Hole Punching to NAT Initiator", "dst", natInitiatorAddress)

	if routeDirection.ProcessCode == entity.TunnelRequestToTRS {
		var trsAddress *net.UDPAddr
		switch h.localIPVersion {
		case entity.TypeLocalIPVersion4:
			trsAddress = &net.UDPAddr{
				IP:   routeDirection.TRSIPv4[:],
				Port: int(routeDirection.TRSPort),
			}
		case entity.TypeLocalIPVersion6, entity.TypeLocalDualStackNetwork:
			trsAddress = &net.UDPAddr{
				IP:   routeDirection.TRSIPv6[:],
				Port: int(routeDirection.TRSPort),
			}
		}

		// to TRS
		generatedPackets = append(generatedPackets,
			&entity.OutPacketQueue{
				Buffer: buffer,
				Addr:   trsAddress,
			})

		config.LogDebug("Generated Hole Punching to TRS", "dst", trsAddress)
	}

	return generatedPackets, nil
}

func (h *handler) handleRouteDirectionToInitiator(packet *entity.BasePacket) ([]*entity.OutPacketQueue, error) {
	config.LogDebug("Handle Route Direction To Initiator")

	routeDirection, err := entity.ParseRouteDirection(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse route direction to responder: %w", err)
	}

	err = h.routeDirectionCache.Put(
		hex.EncodeToString(routeDirection.PathID),
		routeDirection,
		time.Now().Add(60*time.Second).UnixNano())

	if err != nil {
		return nil, fmt.Errorf("failed to put route direction to cache: %w", err)
	}

	endKey, err := entity.GenerateEndKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate end key: %w", err)
	}

	tunnelRequest, err := entity.GenerateTunnelRequest(routeDirection.PathID,
		routeDirection.TemporaryKey,
		endKey,
		routeDirection.ProcessCode,
		routeDirection.BaseHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tunnel request: %w", err)
	}

	// TODO: Socketの選択ロジックを実装する
	soc, err := h.socketStore.HighestPrioritySocket()
	if err != nil {
		return nil, fmt.Errorf("failed to get highest priority socket: %w", err)
	}

	var peerNodeType entity.NodeType
	if routeDirection.HasSameNAT(h.localIPVersion) {
		peerNodeType = entity.NodeTypeLocalResponder
	} else {
		peerNodeType = entity.NodeTypeResponder
	}

	peer, err := entity.GeneratePeer(peerNodeType, routeDirection)
	if err != nil {
		return nil, fmt.Errorf("failed to create new peer: %w", err)
	}

	tunnel := entity.GenerateTunnel(peerNodeType, routeDirection, soc, endKey)

	h.peerRegistry.UpsertPeer(peer)
	if err := h.peerRegistry.AssignTunnel(tunnel); err != nil {
		return nil, fmt.Errorf("failed to assign tunnel to peer: %w", err)
	}

	buffer, err := tunnelRequest.Marshal(routeDirection.TunnelKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tunnel request: %w", err)
	}

	// address determination process
	var sendAddress *net.UDPAddr

	if routeDirection.ProcessCode == entity.TunnelRequestToTRS {
		// via TRS
		switch h.localIPVersion {
		case entity.TypeLocalIPVersion4:
			sendAddress = &net.UDPAddr{
				IP:   routeDirection.TRSIPv4[:],
				Port: int(routeDirection.TRSPort),
			}
		case entity.TypeLocalIPVersion6, entity.TypeLocalDualStackNetwork:
			sendAddress = &net.UDPAddr{
				IP:   routeDirection.TRSIPv6[:],
				Port: int(routeDirection.TRSPort),
			}
		}
	} else {
		// not via TRS
		switch h.localIPVersion {
		case entity.TypeLocalIPVersion4:
			sendAddress = tunnel.Endpointv4()
		case entity.TypeLocalIPVersion6, entity.TypeLocalDualStackNetwork:
			sendAddress = tunnel.Endpointv6()
		}
	}

	config.LogDebug("Generated Tunnel Request",
		"socket_address", soc.LocalAddr(),
		"send_address", sendAddress,
		"transaction_id", tunnelRequest.BaseHeader.TransactionID,
		"path_id", tunnelRequest.BaseHeader.ID,
		"tunnel_request", tunnelRequest)

	return []*entity.OutPacketQueue{{
		Buffer: buffer,
		Addr:   sendAddress,
		Soc:    soc,
	}}, nil
}

func (h *handler) handleTunnelRequestForUDP(packet *entity.BasePacket, peerAddr *net.UDPAddr) ([]*entity.OutPacketQueue, error) {
	config.LogDebug("Handle Tunnel Request For UDP")

	routeDirection, _, _, err := h.routeDirectionCache.Get(hex.EncodeToString(packet.BaseHeader.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to get route direction from cache: %w", err)
	}

	tunnelRequest, err := entity.ParseTunnelRequest(packet, routeDirection.TunnelKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse route direction to responder: %w", err)
	}

	tunnelResponse := entity.GenerateTunnelResponse(tunnelRequest.BaseHeader.ID, tunnelRequest.BaseHeader)

	config.LogDebug("Generated Tunnel Response For UDP",
		"transaction_id", tunnelRequest.BaseHeader.TransactionID,
		"path_id", tunnelRequest.BaseHeader.ID,
		"tunnel_response", tunnelResponse)

	// TODO: Socketの選択ロジックを実装する
	soc, err := h.socketStore.HighestPrioritySocket()
	if err != nil {
		return nil, fmt.Errorf("failed to get highest priority socket: %w", err)
	}

	var peerNodeType entity.NodeType
	if routeDirection.HasSameNAT(h.localIPVersion) {
		peerNodeType = entity.NodeTypeLocalInitiator
	} else {
		peerNodeType = entity.NodeTypeInitiator
	}

	peer, err := entity.GeneratePeer(peerNodeType, &routeDirection)
	if err != nil {
		return nil, fmt.Errorf("failed to create new peer: %w", err)
	}

	tunnel := entity.GenerateTunnel(peerNodeType, &routeDirection, soc, tunnelRequest.EndKey)

	h.peerRegistry.UpsertPeer(peer)
	if err := h.peerRegistry.AssignTunnel(tunnel); err != nil {
		return nil, fmt.Errorf("failed to assign tunnel to peer: %w", err)
	}

	buffer, err := tunnelResponse.Marshal(routeDirection.TunnelKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tunnel request: %w", err)
	}

	return []*entity.OutPacketQueue{{
		Buffer: buffer,
		Addr:   peerAddr,
		Soc:    soc,
	}}, nil
}

func (h *handler) handleTunnelResponseForUDP(packet *entity.BasePacket, peerAddr *net.UDPAddr) (*entity.DNSAnswerQueue, []*entity.OutPacketQueue, error) {
	config.LogDebug("Handle Tunnel Response For UDP")

	routeDirection, _, _, err := h.routeDirectionCache.Get(hex.EncodeToString(packet.BaseHeader.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get route direction from cache: %w", err)
	}

	// get the DNS identifier from the DNS cache
	dstPort, id := h.dnsCache.Get(routeDirection.BaseHeader.TransactionID)

	// TODO: transactionIDの型をuint32に統一
	dnsAnswer := entity.DNSAnswerQueue{
		TransactionID: routeDirection.BaseHeader.TransactionID,
		FQDN:          string(routeDirection.ResponderFQDN),
		VirtualIPv4:   routeDirection.ResponderVirtualIPv4[:],
		VirtualIPv6:   routeDirection.ResponderVirtualIPv6[:],
		DNSIdentifier: entity.DNSIdentifier{
			ID:   id,
			Port: dstPort,
		},
	}

	tunnel, ok := h.peerRegistry.GetTunnelByPathID(routeDirection.PathID)
	if !ok {
		return nil, nil, fmt.Errorf("tunnel not found for path ID: %s", hex.EncodeToString(routeDirection.PathID))
	}

	if !tunnel.ViaTRS() {
		return &dnsAnswer, nil, nil
	}
	if !h.routeOptimizationMode {
		return &dnsAnswer, nil, nil
	}

	// route optimization process
	holePunchingForOptimization := entity.GenerateHolePunchingForOptimization(routeDirection.PathID, h.baseHeader)
	config.LogDebug("Generated Hole Punching For Optimization")

	buffer, err := holePunchingForOptimization.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal Hole Punching: %w", err)
	}

	var sendAddress *net.UDPAddr
	switch h.localIPVersion {
	case entity.TypeLocalIPVersion4:
		sendAddress = tunnel.Endpointv4()
	case entity.TypeLocalIPVersion6, entity.TypeLocalDualStackNetwork:
		sendAddress = tunnel.Endpointv6()
	}

	return &dnsAnswer, []*entity.OutPacketQueue{{
		Buffer: buffer,
		Addr:   sendAddress,
	}}, nil
}

func (h *handler) handleHolePunching(packet *entity.BasePacket, peerAddr *net.UDPAddr) ([]*entity.OutPacketQueue, error) {
	config.LogDebug("handle Hole Punching")

	// TODO: Hole Punchingのハンドル処理(peerの保存)
	return nil, nil
}

func (h *handler) handleHolePunchingForOptimization(packet *entity.BasePacket, peerAddr *net.UDPAddr) ([]*entity.OutPacketQueue, error) {
	config.LogDebug("Handle Hole Punching For Optimization")

	tunnel, ok := h.peerRegistry.GetTunnelByPathID(packet.BaseHeader.ID)
	if !ok {
		return nil, fmt.Errorf("tunnel not found for path ID: %s", hex.EncodeToString(packet.BaseHeader.ID))
	}

	if !tunnel.ViaTRS() {
		config.LogDebug("Skip handling Hole Punching for Optimization because the tunnel is not via TRS")
		return nil, nil
	}

	if err := tunnel.UpdateEndpointAddr(peerAddr); err != nil {
		return nil, fmt.Errorf("failed to update tunnel endpoint address: %w", err)
	}

	tunnel.RouteOptimized()

	ackForOptimization := entity.GenerateACKForOptimization(packet.BaseHeader.ID, packet.BaseHeader)
	config.LogDebug("Generated ACK For Optimization")

	buffer, err := ackForOptimization.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ACK: %w", err)
	}

	return []*entity.OutPacketQueue{{
		Buffer: buffer,
		Addr:   peerAddr,
		Soc:    tunnel.Socket(),
	}}, nil
}

func (h *handler) handleAckForOptimization(packet *entity.BasePacket, peerAddr *net.UDPAddr) ([]*entity.OutPacketQueue, error) {
	config.LogDebug("Handle ACK For Optimization")

	tunnel, ok := h.peerRegistry.GetTunnelByPathID(packet.BaseHeader.ID)
	if !ok {
		return nil, fmt.Errorf("tunnel not found for path ID: %s", hex.EncodeToString(packet.BaseHeader.ID))
	}

	if !tunnel.ViaTRS() {
		config.LogDebug("Skip handling Hole Punching for Optimization because the tunnel is not via TRS")
		return nil, nil
	}

	if err := tunnel.UpdateEndpointAddr(peerAddr); err != nil {
		return nil, fmt.Errorf("failed to update tunnel endpoint address: %w", err)
	}

	tunnel.RouteOptimized()
	config.LogDebug("Route optimized for tunnel", "path_id", hex.EncodeToString(tunnel.PathID()))

	return nil, nil
}

func (h *handler) isKeepAliveAckFromNS(peerAddr *net.UDPAddr) bool {
	if peerAddr == nil {
		config.LogErr("peerAddr is nil in handleKeepAliveAck")
		return false
	}
	peerIPAddr, ok := netip.AddrFromSlice(peerAddr.IP)
	if !ok {
		config.LogErr("failed to parse peerAddr IP", "peerAddr", peerAddr.IP.String())
		return false
	}

	//Check if the Peer address matches the NS address.
	if !h.nsAddress.IsNSAddr(peerIPAddr) {
		config.LogErr("Received KeepAlive Ack from unexpected address", "peerAddr", peerAddr.String())
		return false
	}

	return true
}
