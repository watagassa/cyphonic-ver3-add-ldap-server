package entity

import (
	"bytes"
	"fmt"
)

// ACK is ACK of a CYPHONIC packet.
type ACK struct {
	BaseHeader BaseHeader
	HMAC       HMAC
}

// GenerateACK generates ACK packet.
func GenerateACK(pathID ID, baseHeader BaseHeader) *ACK {
	baseHeader.SerializeBaseHeader(TypeClassAck, StatusClassSuccess, 0)
	baseHeader.ID = pathID

	return &ACK{
		BaseHeader: baseHeader,
		HMAC:       make(HMAC, HMACLength),
	}
}

func GenerateKeepAliveACK(pathID ID, baseHeader BaseHeader) *ACK {
	baseHeader.SerializeBaseHeader(TypeClassKeepAliveAck, StatusClassSuccess, 0)
	baseHeader.ID = pathID

	return &ACK{
		BaseHeader: baseHeader,
		HMAC:       make(HMAC, HMACLength),
	}
}

// Marshal returns binary data.
func (ack *ACK) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	baseHeaderBuf := ack.BaseHeader.Serialize()
	if _, err := buffer.Write(baseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	buffer.Write(ack.HMAC)

	return buffer.Bytes(), nil
}
