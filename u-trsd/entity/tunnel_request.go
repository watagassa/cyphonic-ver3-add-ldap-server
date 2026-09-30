package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// TunnelRequestLength is the Length of TunnelRequest packet.
const TunnelRequestLength = BaseHeaderLength + 4 + HMACLength

// TunnelRequest is Tunnel Request of a CYPHONIC packet.
type TunnelRequest struct {
	BaseHeader   BaseHeader
	EndKeyLength uint16
	Padding      uint16
	EndKey       []byte
	HMAC         HMAC
}

// Marshal serializes the TunnelRequestPacket into a byte slice.
func (t *TunnelRequest) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	baseHeaderBuf := t.BaseHeader.Serialize()
	if _, err := buffer.Write(baseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	tunnelRequestFields := []interface{}{
		t.EndKeyLength,
		t.Padding,
	}

	for _, field := range tunnelRequestFields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to marshal tunnel request: %w", err)
		}
	}

	buffer.Write(t.EndKey)
	buffer.Write(t.HMAC)

	return buffer.Bytes(), nil
}

// ParseTunnelRequest parses a TunnelRequest from a BasePacket.
func ParseTunnelRequest(b *BasePacket, key []byte, keyCipherType uint16, keyLength uint16) (TunnelRequest, error) {
	if len(b.Payload) < TunnelRequestLength-BaseHeaderLength {
		return TunnelRequest{}, ErrInvalidPacket
	}

	// Decrypt the payload
	encryptedLength := len(b.Payload) - HMACLength
	packet, err := DecryptPacket(b.Payload[0:encryptedLength], key, keyCipherType, keyLength)
	if err != nil {
		return TunnelRequest{}, fmt.Errorf("failed to decrypt Tunnel Request: %w", err)
	}

	// Check Packet Length
	if len(packet) != int(b.BaseHeader.MessageLength) {
		fmt.Println("b.BaseHeader.MessageLength-HMACLength: ", b.BaseHeader.MessageLength-HMACLength)
		return TunnelRequest{}, ErrInvalidPacket
	}

	endkeyLength := binary.BigEndian.Uint16(packet[0:2])

	return TunnelRequest{
		BaseHeader:   b.BaseHeader,
		EndKeyLength: endkeyLength,
		Padding:      binary.BigEndian.Uint16(packet[2:4]),
		EndKey:       packet[4 : 4+endkeyLength],
		HMAC:         b.Payload[encryptedLength : encryptedLength+HMACLength],
	}, nil
}
