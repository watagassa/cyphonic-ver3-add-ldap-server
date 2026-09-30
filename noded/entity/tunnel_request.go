package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

// TunnelRequest is Tunnel Request of a CYPHONIC packet.
type TunnelRequest struct {
	BaseHeader   BaseHeader
	EndKeyLength uint16
	Padding      uint16
	EndKey       []byte
	HMAC         HMAC
}

// Payload length in TunnelRequest.
const (
	// EndKeyLengthSize EndKeyLength lengths (bytes).
	EndKeyLengthSize = 2

	// TunnelRequestPadding padding lengths (bytes).
	TunnelRequestPadding = 2
)

const BaseTunnelRequestLength = EndKeyLengthSize + TunnelRequestPadding

func GenerateTunnelRequest(pathID ID, tempKey, endKey []byte, process ProcessCodeClass, baseHeader BaseHeader) (*TunnelRequest, error) {
	// TODO: TRS対応
	// if process == TunnelRequestToTRS {
	// 	encryptEndKey(tempKey)
	// }

	endKeyLength := len(endKey)
	TunnelRequestLength := BaseTunnelRequestLength + endKeyLength
	baseHeader.SerializeBaseHeader(TypeClassTunnelRequestForUDP, StatusClassSuccess, uint16(TunnelRequestLength))
	baseHeader.ID = pathID

	return &TunnelRequest{
		BaseHeader:   baseHeader,
		EndKeyLength: uint16(endKeyLength),
		Padding:      0,
		EndKey:       endKey,
		HMAC: HMAC{
			HMAC: make([]byte, HMACLength),
		},
	}, nil
}

// GenerateEndKey generates an end key to encrypt the contents of the packet.
func GenerateEndKey() ([]byte, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID for EndKey: %w", err)
	}

	ub, err := u.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("for EndKey: failed encode to binary format: %w", err)
	}

	return ub, nil
}

// Marshal TODO: Keyによる暗号化
func (t *TunnelRequest) Marshal(key []byte) ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	baseHeaderBuf := t.BaseHeader.Serialize()
	if _, err := buffer.Write(baseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	tunnelRequestFields := []interface{}{
		t.EndKeyLength, t.Padding,
	}

	for _, field := range tunnelRequestFields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to marshal tunnel request: %w", err)
		}
	}

	buffer.Write(t.EndKey)
	buffer.Write(t.HMAC.HMAC)

	return buffer.Bytes(), nil
}

func ParseTunnelRequest(packet *BasePacket, key []byte) (*TunnelRequest, error) {
	// TODO: keyによるDecrypt処理
	if len(packet.Payload) < BaseTunnelRequestLength {
		return nil, fmt.Errorf("invalid packet length: %d", len(packet.Payload))
	}

	tunnelRequest := &TunnelRequest{
		BaseHeader:   packet.BaseHeader,
		EndKeyLength: binary.BigEndian.Uint16(packet.Payload[0:2]),
		Padding:      binary.BigEndian.Uint16(packet.Payload[2:4]),
	}

	tunnelRequest.EndKey = packet.Payload[4 : 4+int(tunnelRequest.EndKeyLength)]

	// Validate the lengths of EndKey
	if len(tunnelRequest.EndKey) != int(tunnelRequest.EndKeyLength) {
		return nil, fmt.Errorf("invalid end key length: expected %d, got %d", tunnelRequest.EndKeyLength, len(tunnelRequest.EndKey))
	}

	// Check if the payload length matches the expected length
	expectedLength := BaseTunnelRequestLength + int(tunnelRequest.EndKeyLength)
	// (BaseHeaderLength + HMACLength)

	if len(packet.Payload) != expectedLength {
		return nil, fmt.Errorf("invalid packet length: expected %d, got %d", expectedLength, len(packet.Payload))
	}
	return tunnelRequest, nil
}
