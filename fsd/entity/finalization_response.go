package entity

import (
	"bytes"
	"fmt"
)

// FinalizationResponse is Provision Response of a CYPHONIC packet.
type FinalizationResponse struct {
	BaseHeader BaseHeader
}

// BaseFinalizationResponseLength is length of (base) body part of FinalizationResponse.
// Also, since the length of FQDN is variable length, when
// marshaling, calculate and add.
const BaseFinalizationResponseLength = BaseHeaderLength

// SerializeFinalizationResponsePacketLength sets the length of the Finalization Response packet.
func (response *FinalizationResponse) SerializeFinalizationResponsePacketLength() {
	response.BaseHeader.MessageLength = BaseFinalizationResponseLength
}

func (response *FinalizationResponse) ConvertBytes() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	BaseHeaderBuf := response.BaseHeader.All()
	if _, err := buffer.Write(BaseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	return buffer.Bytes(), nil
}
