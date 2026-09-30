package entity

import (
	"encoding/binary"
	"errors"
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
func ParseDirectionRequest(packet *BasePacket, commonKey []byte) (*DirectionRequest, error) {
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

func GenerateDirectionRequest(prevDirectionRequest *DirectionRequest) *DirectionRequest {
	directionRequest := &DirectionRequest{
		BaseHeader: BaseHeader{
			TransactionID:  prevDirectionRequest.BaseHeader.TransactionID,
			Version:        prevDirectionRequest.BaseHeader.Version,
			Type:           TypeClassDirectionRequest,
			Status:         0,
			Count:          prevDirectionRequest.BaseHeader.Count + 1,
			SequenceNumber: prevDirectionRequest.BaseHeader.SequenceNumber + 1,
			MessageLength:  BaseRegistrationResponseLength + HMACLength,
			Token:          0,
			NextOpt:        0,
			ID:             prevDirectionRequest.BaseHeader.ID,
		},
		ApplicationID:        prevDirectionRequest.ApplicationID,
		PathID:               prevDirectionRequest.PathID,
		TunnelConnectionType: prevDirectionRequest.TunnelConnectionType,
		FQDNLength:           uint16(len(prevDirectionRequest.FQDN)),
		FQDN:                 prevDirectionRequest.FQDN,
	}

	return directionRequest
}

func (request *DirectionRequest) ConvertBytes() ([]byte, error) {
	buf := make([]byte, BaseDirectionRequestLength)

	binary.BigEndian.PutUint32(buf[0:4], request.BaseHeader.TransactionID)
	buf[4] = request.BaseHeader.Version
	buf[5] = byte(request.BaseHeader.Type)
	buf[6] = byte(request.BaseHeader.Status)
	buf[7] = request.BaseHeader.Count
	binary.BigEndian.PutUint32(buf[8:12], request.BaseHeader.SequenceNumber)
	binary.BigEndian.PutUint16(buf[12:14], request.BaseHeader.MessageLength)
	buf[14] = request.BaseHeader.Token
	buf[15] = request.BaseHeader.NextOpt
	binary.BigEndian.PutUint64(buf[16:24], binary.BigEndian.Uint64(request.BaseHeader.ID[0:8]))
	binary.BigEndian.PutUint64(buf[24:32], binary.BigEndian.Uint64(request.BaseHeader.ID[8:16]))
	// Registration Request Body
	binary.BigEndian.PutUint64(buf[32:40], binary.BigEndian.Uint64(request.ApplicationID[0:8]))
	binary.BigEndian.PutUint64(buf[40:48], binary.BigEndian.Uint64(request.ApplicationID[8:16]))
	binary.BigEndian.PutUint64(buf[48:56], binary.BigEndian.Uint64(request.PathID[0:8]))
	binary.BigEndian.PutUint64(buf[56:64], binary.BigEndian.Uint64(request.PathID[8:16]))
	binary.BigEndian.PutUint16(buf[64:66], request.TunnelConnectionType)
	binary.BigEndian.PutUint16(buf[66:68], request.FQDNLength)

	if len(request.FQDN) != int(request.FQDNLength) {
		return nil, errors.New("FQDN does not match FQDN length")
	}
	buf = append(buf, request.FQDN...)

	return buf, nil
}
