package entity

import (
	"encoding/binary"
	"fmt"
)

type DirectionRequest struct {
	BaseHeader           BaseHeader
	ApplicationID        ID
	PathID               ID
	TunnelConnectionType uint16
	FQDNLength           uint16
	FQDN                 []byte
}

// Payload length in DirectionRequest.
const (
	// ApplicationIDSize ApplicationID lengths (bytes).
	ApplicationIDSize = 16
	// PathIDSize PathID lengths (bytes).
	PathIDSize = 16
	// TunnelConnectionTypeSize TunnelConnectionType lengths (bytes).
	TunnelConnectionTypeSize = 2
	// FQDNLengthSize FQDNLength lengths (bytes).
	FQDNLengthSize = 2
)

const BaseDirectionRequestLength = BaseHeaderLength +
	ApplicationIDSize +
	PathIDSize +
	TunnelConnectionTypeSize +
	FQDNLengthSize

type TunnelConnectionType uint16

const (
	TunnelConnectionUDP TunnelConnectionType = iota
	TunnelConnectionQUIC
	TunnelConnectionBoth
)

// ParseDirectionRequest parses a DirectionRequest packet.
func ParseDirectionRequest(packet *BasePacket) (*DirectionRequest, error) {
	// Check if the payload length is more than the minimum packetsize
	if len(packet.Payload) < (BaseDirectionRequestLength - BaseHeaderLength) {
		return nil, ErrInvalidPacket
	}

	directionRequest := &DirectionRequest{
		BaseHeader:           packet.BaseHeader,
		ApplicationID:        packet.Payload[0:NodeIDlength],
		PathID:               packet.Payload[NodeIDlength : NodeIDlength+PathIDlen],
		TunnelConnectionType: binary.BigEndian.Uint16(packet.Payload[NodeIDlength+PathIDlen : NodeIDlength+PathIDlen+2]),
		FQDNLength:           binary.BigEndian.Uint16(packet.Payload[NodeIDlength+PathIDlen+2 : NodeIDlength+PathIDlen+4]),
		FQDN:                 packet.Payload[NodeIDlength+PathIDlen+4 : NodeIDlength+PathIDlen+4+binary.BigEndian.Uint16(packet.Payload[NodeIDlength+PathIDlen+2:])],
	}

	// Check if the payload length matches the expected length
	expectedLength := BaseDirectionRequestLength +
		int(directionRequest.FQDNLength) -
		BaseHeaderLength

	if len(packet.Payload) != expectedLength {
		return nil, fmt.Errorf("invalid packet length: expected %d, got %d", expectedLength, len(packet.Payload))
	}

	return directionRequest, nil
}
