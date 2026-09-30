package entity

import (
	"bytes"
	"fmt"
)

// FinalizationResponse is Finalization Response of a CYPHONIC packet.
type FinalizationResponse struct {
	BaseHeader BaseHeader
}

// BaseFinalizationResponseLength is length of (base) body part of FinalizationResponse.
// Also, since the length of FQDN is variable length, when
// marshaling, calculate and add.
const BaseFinalizationResponseLength = BaseHeaderLength

// SerializeFinalizationResponsePacketLength SerializeLoginResponsePacketLength sets the length of the Login Response packet.
func (freq *FinalizationResponse) SerializeFinalizationResponsePacketLength() {
	freq.BaseHeader.MessageLength = BaseFinalizationResponseLength
}

func GenerateFinalizationResponse(baseHeader BaseHeader) *FinalizationResponse {
	baseHeader.SerializeBaseHeader(TypeClassAck, StatusClassSuccess, uint16(BaseFinalizationResponseLength))

	return &FinalizationResponse{
		BaseHeader: baseHeader,
	}
}

func (freq *FinalizationResponse) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	BaseHeaderBuf := freq.BaseHeader.All()
	if _, err := buffer.Write(BaseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	return buffer.Bytes(), nil
}

func UnmarshalFinalizationResponse(buffer []byte) (*FinalizationResponse, error) {
	if len(buffer) < BaseFinalizationResponseLength {
		return nil, fmt.Errorf("%w: %d", ErrInvalidBufferLength, len(buffer))
	}

	baseHeader := UnmarshalBaseHeader(buffer[:BaseHeaderLength])

	return &FinalizationResponse{
		BaseHeader: *baseHeader,
	}, nil
}
