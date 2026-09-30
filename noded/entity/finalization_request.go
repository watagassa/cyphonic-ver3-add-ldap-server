package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// FinalizationRequest is Finalization Request of a CYPHONIC packet.
type FinalizationRequest struct {
	BaseHeader        BaseHeader
	AccessTokenLength uint16
	Padding           uint16
	AccessToken       []byte
}

const BaseFinalizationRequestLength = BaseHeaderLength + AccessTokenLengthSize + PaddingSize

func GenerateFinalizationRequest(baseHeader BaseHeader, accessToken []byte) *FinalizationRequest {
	BaseFinalizationRequestLength := BaseFinalizationRequestLength + len(accessToken)
	baseHeader.SerializeBaseHeader(TypeClassFinalizationRequest, StatusClassSuccess, uint16(BaseFinalizationRequestLength))

	return &FinalizationRequest{
		BaseHeader:        baseHeader,
		AccessTokenLength: uint16(len(accessToken)),
		Padding:           0,
		AccessToken:       accessToken,
	}
}

func (freq FinalizationRequest) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	BaseHeaderBuf := freq.BaseHeader.All()
	if _, err := buffer.Write(BaseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	fields := []interface{}{
		freq.AccessTokenLength,
		freq.Padding,
		freq.AccessToken,
	}
	for _, field := range fields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to write field: %w", err)
		}
	}

	return buffer.Bytes(), nil
}
