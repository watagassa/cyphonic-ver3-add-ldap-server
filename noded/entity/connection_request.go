package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ConnectionRequest struct {
	BaseHeader        BaseHeader
	AccessTokenLength uint16
	Padding           uint16
	AccessToken       []byte
}

const BaseConnectionRequestLength = BaseHeaderLength + AccessTokenLengthSize + PaddingSize

func GenerateConnectionRequest(baseHeader BaseHeader, accessToken []byte) *ConnectionRequest {
	BaseConnectionRequestLength := BaseConnectionRequestLength + len(accessToken)
	baseHeader.SerializeBaseHeader(TypeClassConnectionRequest, StatusClassSuccess, uint16(BaseConnectionRequestLength))

	return &ConnectionRequest{
		BaseHeader:        baseHeader,
		AccessTokenLength: uint16(len(accessToken)),
		Padding:           0,
		AccessToken:       accessToken,
	}
}

func (creq ConnectionRequest) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	BaseHeaderBuf := creq.BaseHeader.All()
	if _, err := buffer.Write(BaseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	fields := []interface{}{
		creq.AccessTokenLength,
		creq.Padding,
		creq.AccessToken,
	}
	for _, field := range fields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to write field: %w", err)
		}
	}

	return buffer.Bytes(), nil
}
