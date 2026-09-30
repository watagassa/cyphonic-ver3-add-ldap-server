package entity

import (
	"encoding/binary"
	"errors"
)

// ErrInvalidPacket is returned when a packet is invalid.
var ErrInvalidPacket = errors.New("invalid packet")

// BasePacket is a convenience struct which implements the PacketData interface.
type BasePacket struct {
	BaseHeader BaseHeader // BaseHeader is the set of bytes that make up this packet.
	Payload    []byte     // Payload is the payload contained within this packet.
}

// BaseHeaderAndPayload returns the bytes of the packet, including the BaseHeader and Payload.
func (b *BasePacket) BaseHeaderAndPayload() []byte {
	return append(b.BaseHeader.Serialize(), b.Payload...)
}

// PacketBaseHeader returns the bytes of the packet.
func (b *BasePacket) PacketBaseHeader() BaseHeader { return b.BaseHeader }

// PacketPayload returns the payload contained within the packet.
func (b *BasePacket) PacketPayload() []byte { return b.Payload }

// Marshal serializes the BasePacket into a byte slice.
func (b *BasePacket) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	data = append(data, b.BaseHeader.Serialize()...)
	data = append(data, b.Payload...)
	return data, nil
}

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

// ParsePlainPacket parses a byte slice into a BasePacket structure.
func ParsePlainPacket(data []byte) (*BasePacket, error) {
	if len(data) < BaseHeaderLength {
		return nil, ErrInvalidPacket
	}

	packet := &BasePacket{
		BaseHeader: parseBaseHeader(data),
		Payload:    data[BaseHeaderLength:],
	}

	return packet, nil
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
