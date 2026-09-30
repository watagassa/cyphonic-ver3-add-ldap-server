package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

type DirectionRequest struct {
	BaseHeader           BaseHeader
	ApplicationID        ID
	PathID               ID
	TunnelConnectionType TunnelConnectionType
	FQDNLength           uint16
	FQDN                 []byte
	HMAC                 HMAC
}

// Payload length in DirectionRequest.
const (
	// ApplicationIDSize ApplicationID lengths (bytes).
	ApplicationIDSize = 16
	// PathIDSize PathID lengths (bytes).
	PathIDSize = 16
	// TunnelConnectionTypeSize TunnelConnectionType lengths (bytes).
	TunnelConnectionTypeSize = 2
	// FQDNLengthOfDirectionRequestSize FQDNLength lengths (bytes).
	FQDNLengthOfDirectionRequestSize = 2
)

const BaseDirectionRequestLength = BaseHeaderLength +
	ApplicationIDSize +
	PathIDSize +
	TunnelConnectionTypeSize +
	FQDNLengthOfDirectionRequestSize +
	HMACLength

type TunnelConnectionType uint16

const (
	TunnelConnectionUDP TunnelConnectionType = iota
	TunnelConnectionQUIC
	TunnelConnectionBoth
)

func GenerateDirectionRequest(baseHeader BaseHeader, inFQDN, rnFQDN []byte) (*DirectionRequest, error) {
	DirectionRequestLength := BaseDirectionRequestLength + len(rnFQDN)
	baseHeader.SerializeBaseHeader(TypeClassDirectionRequest, StatusClassSuccess, uint16(DirectionRequestLength))

	pathID, err := GeneratePathID(inFQDN, rnFQDN)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PathID: %w", err)
	}

	// TODO: implement ApplicationID and TunnelConnectionType
	return &DirectionRequest{
		BaseHeader: baseHeader,
		ApplicationID: []byte{
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		},
		PathID:               pathID,
		TunnelConnectionType: TunnelConnectionUDP,
		FQDNLength:           uint16(len(rnFQDN)),
		FQDN:                 rnFQDN,
		HMAC: HMAC{
			HMAC: make([]byte, HMACLength),
		},
	}, nil
}

func GeneratePathID(inFQDN, rnFQDN []byte) (b []byte, err error) {
	seed := append(inFQDN, rnFQDN...)

	pathID := uuid.NewSHA1(uuid.NameSpaceDNS, seed)

	if b, err = pathID.MarshalBinary(); err != nil {
		return nil, err
	}

	return b, nil
}

func (dreq DirectionRequest) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	BaseHeaderBuf := dreq.BaseHeader.All()
	if _, err := buffer.Write(BaseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	fields := []interface{}{
		dreq.ApplicationID,
		dreq.PathID,
		dreq.TunnelConnectionType,
		dreq.FQDNLength,
		dreq.FQDN,
		dreq.HMAC.HMAC,
	}
	for _, field := range fields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to write field: %w", err)
		}
	}

	return buffer.Bytes(), nil
}
