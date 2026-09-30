package entity

import (
	"encoding/binary"
	"errors"
)

var ErrInvalidPacket = errors.New("invalid packet")

// BasePacket is a convenience struct which implements the PacketData and
// PacketPayload and PacketHMAC functions of the Packet interface.
type BasePacket struct {
	// BaseHeader is the set of bytes that make up this packet.
	BaseHeader BaseHeader
	// Payload is the set of bytes contained by (but not part of) this
	// Layer.  Again, to take Ethernet as an example, this would be the
	// set of bytes encapsulated by the Ethernet protocol.
	Payload []byte
	// HMAC is Hash-based Message Authentication Code of a CYPHONIC packet.
	HMAC []byte
}

func (b *BasePacket) BaseHeaderAndPayload() []byte {
	return append(b.BaseHeader.All(), b.Payload...)
}

// PacketBaseHeader returns the bytes of the packet.
func (b *BasePacket) PacketBaseHeader() BaseHeader { return b.BaseHeader }

// PacketPayload returns the payload contained within the packet.
func (b *BasePacket) PacketPayload() []byte { return b.Payload }

// PacketHMAC returns the HMAC contained within the packet.
func (b *BasePacket) PacketHMAC() []byte { return b.HMAC }

// NodeID offset length.
const (
	MessageOffsetNodeID = 16
	NodeIDlength        = 16
)

// PathID offset length.
const (
	MessageOffsetPathID = 16
	PathIDlen           = 16
)

// MessageOffsetPacketType offset length.
const MessageOffsetPacketType = 6

// ParsePacket parses the raw data of a UDP packet and determines its type.
func ParsePacket(data []byte) (*BasePacket, error) {
	if len(data) < BaseHeaderLength+HMACLength {
		return nil, ErrInvalidPacket
	}

	return &BasePacket{
		BaseHeader: parseBaseHeader(data),
		Payload:    data[BaseHeaderLength : len(data)-HMACLength], //Payload部分はそのまま
		HMAC:       data[len(data)-HMACLength:],
	}, nil
}

// ChangeNoHMAC append the HMAC field to the Payload field
func (b *BasePacket) ChangeNoHMAC() error {
	b.Payload = append(b.Payload, b.HMAC...)
	b.HMAC = nil
	return nil
}

func parseBaseHeader(data []byte) BaseHeader {
	return BaseHeader{
		TransactionID:  binary.BigEndian.Uint32(data[0:4]),
		Version:        data[4],
		Type:           TypeClass(data[5]),
		Status:         StatusClass(data[6]),
		Count:          data[7],
		SequenceNumber: binary.BigEndian.Uint32(data[8:12]),
		MessageLength:  binary.BigEndian.Uint16(data[12:14]),
		Token:          data[14],
		NextOpt:        data[15],
		ID:             data[16:32],
	}
}
