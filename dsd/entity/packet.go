package entity

import (
	"encoding/binary"
	"errors"
)

var ErrInvalidPacket = errors.New("invalid packet")

// BasePacket is a convenience struct which implements the PacketData and
// PacketPayload functions of the Packet interface.
type BasePacket struct {
	// BaseHeader is the set of bytes that make up this packet.
	BaseHeader BaseHeader
	// Payload is the set of bytes contained by (but not part of) this
	// Layer.  Again, to take Ethernet as an example, this would be the
	// set of bytes encapsulated by the Ethernet protocol.
	Payload []byte
}

func (b *BasePacket) BaseHeaderAndPayload() []byte {
	return append(b.BaseHeader.Serialize(), b.Payload...)
}

// PacketBaseHeader returns the bytes of the packet.
func (b *BasePacket) PacketBaseHeader() BaseHeader { return b.BaseHeader }

// PacketPayload returns the payload contained within the packet.
func (b *BasePacket) PacketPayload() []byte { return b.Payload }

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
