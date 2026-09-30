package entity

import (
	"bytes"
	"fmt"
)

// KeepAlive is keep alive of a CYPHONIC packet.
type KeepAlive struct {
	BaseHeader BaseHeader
	HMAC       HMAC
}

const BaseKeepAliveLength = BaseHeaderLength + HMACLength

func GenerateKeepAlive(baseHeader BaseHeader) *KeepAlive {
	baseHeader.SerializeBaseHeader(TypeClassKeepAlive, StatusClassSuccess, uint16(BaseKeepAliveLength))

	return &KeepAlive{
		BaseHeader: baseHeader,
		HMAC: HMAC{
			HMAC: make([]byte, HMACLength),
		},
	}
}

func (ka KeepAlive) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	baseHeaderBuf := ka.BaseHeader.Serialize()
	if _, err := buffer.Write(baseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	buffer.Write(ka.HMAC.HMAC)

	return buffer.Bytes(), nil
}
