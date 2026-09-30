package entity

import (
	"bytes"
	"fmt"
)

// HolePunching is HolePunching of a CYPHONIC packet.
type HolePunching struct {
	BaseHeader BaseHeader
	HMAC       HMAC
}

// GenerateHolePunching generates HolePunching packet.
func GenerateHolePunching(pathID ID, baseHeader BaseHeader) *HolePunching {
	baseHeader.SerializeBaseHeader(TypeClassHolePunching, StatusClassSuccess, 0)
	baseHeader.ID = pathID

	return &HolePunching{
		BaseHeader: baseHeader,
		HMAC:       HMAC{HMAC: make([]byte, HMACLength)},
	}
}

// GenerateHolePunchingForOptimization generates HolePunching packet for optimization.
func GenerateHolePunchingForOptimization(pathID ID, baseHeader BaseHeader) *HolePunching {
	baseHeader.SerializeBaseHeader(TypeClassHolePunchingForOptimization, StatusClassSuccess, 0)
	baseHeader.ID = pathID

	return &HolePunching{
		BaseHeader: baseHeader,
		HMAC:       HMAC{HMAC: make([]byte, HMACLength)},
	}
}

// Marshal returns binary data.
func (hp *HolePunching) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	baseHeaderBuf := hp.BaseHeader.Serialize()
	if _, err := buffer.Write(baseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	buffer.Write(hp.HMAC.HMAC)

	return buffer.Bytes(), nil
}
