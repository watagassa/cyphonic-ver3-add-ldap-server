package entity

import (
	"bytes"
	"fmt"
)

// TunnelResponse is Tunnel Response of a CYPHONIC packet.
type TunnelResponse struct {
	BaseHeader BaseHeader
	HMAC       HMAC
}

const BaseTunnelResponseLength = 0

func GenerateTunnelResponse(pathID ID, baseHeader BaseHeader) *TunnelResponse {
	baseHeader.SerializeBaseHeader(TypeClassTunnelResponseForUDP, StatusClassSuccess, uint16(BaseTunnelResponseLength))
	baseHeader.ID = pathID

	return &TunnelResponse{
		BaseHeader: baseHeader,
		HMAC: HMAC{
			HMAC: make([]byte, HMACLength),
		},
	}
}

// Marshal TODO: Keyによる暗号化
func (t *TunnelResponse) Marshal(key []byte) ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	baseHeaderBuf := t.BaseHeader.Serialize()
	if _, err := buffer.Write(baseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	buffer.Write(t.HMAC.HMAC)

	return buffer.Bytes(), nil
}

func ParseTunnelResponse(packet *BasePacket) *TunnelResponse {
	return &TunnelResponse{
		BaseHeader: packet.BaseHeader,
		HMAC:       packet.HMAC,
	}
}
