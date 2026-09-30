package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ProvisionRequest struct {
	BaseHeader        BaseHeader
	DesiredFQDNLength uint16
	AccessTokenLength uint16
	DesiredFQDN       []byte
	AccessToken       []byte
}

const (
	DesiredFQDNLengthSize = 2
)

const BaseProvisionRequestLength = BaseHeaderLength + DesiredFQDNLengthSize + AccessTokenLengthSize

func GenerateProvisionRequest(baseHeader BaseHeader, desiredFQDN, accessToken []byte) *ProvisionRequest {
	BaseProvisionRequestLength := BaseProvisionRequestLength + len(desiredFQDN) + len(accessToken)
	baseHeader.SerializeBaseHeader(TypeClassProvisionRequest, StatusClassSuccess, uint16(BaseProvisionRequestLength))

	return &ProvisionRequest{
		BaseHeader:        baseHeader,
		DesiredFQDNLength: uint16(len(desiredFQDN)),
		AccessTokenLength: uint16(len(accessToken)),
		DesiredFQDN:       desiredFQDN,
		AccessToken:       accessToken,
	}
}

func (preq ProvisionRequest) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	BaseHeaderBuf := preq.BaseHeader.All()
	if _, err := buffer.Write(BaseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	fields := []interface{}{
		preq.DesiredFQDNLength,
		preq.AccessTokenLength,
		preq.DesiredFQDN,
		preq.AccessToken,
	}
	for _, field := range fields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to write field: %w", err)
		}
	}

	return buffer.Bytes(), nil
}
