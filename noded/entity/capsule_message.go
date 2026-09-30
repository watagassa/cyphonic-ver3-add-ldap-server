package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// CapsuleMessage is Capsule Message of a CYPHONIC packet.
type CapsuleMessage struct {
	BaseHeader    BaseHeader
	PayloadLength uint16
	Padding       uint16
	Payload       []byte
	HMAC          HMAC
}

// Payload length in CapsuleMessage.
const (
	// PayloadLength EndKeyLength lengths (bytes).
	PayloadLength = 2

	// CapsuleMessagePaddingLength padding lengths (bytes).
	CapsuleMessagePaddingLength = 2
)

const BaseCapsuleMessageLength = PayloadLength + CapsuleMessagePaddingLength

func GenerateCapsuleMessage(pathID []byte, transactionID uint32, rawData []byte, packetType TypeClass) (*CapsuleMessage, error) {
	baseHeader, err := GenerateBaseHeader(packetType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base header: %w", err)
	}

	baseHeader.ID = pathID
	baseHeader.TransactionID = transactionID
	baseHeader.MessageLength = uint16(BaseHeaderLength + BaseCapsuleMessageLength + len(rawData) + HMACLength)

	return &CapsuleMessage{
		BaseHeader:    *baseHeader,
		PayloadLength: uint16(len(rawData)),
		Padding:       0,
		Payload:       rawData,
		HMAC: HMAC{
			HMAC: make([]byte, HMACLength)},
	}, nil
}

func (cm CapsuleMessage) Marshal(key []byte) ([]byte, error) {
	// TODO: 暗号化
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)
	planeData := make([]byte, 0)
	planeBuffer := bytes.NewBuffer(planeData)

	fields := []interface{}{
		cm.PayloadLength,
		cm.Padding,
		cm.Payload,
	}
	for _, field := range fields {
		if err := binary.Write(planeBuffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to write field: %w", err)
		}
	}

	cipherBuffer, err := EncryptPacket(planeBuffer.Bytes(), key, uint16(CipherTypeClassAES256CBC), uint16(len(key)))
	if err != nil {
		return nil, err
	}

	BaseHeaderBuf := cm.BaseHeader.All()
	if _, err = buffer.Write(BaseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	buffer.Write(cipherBuffer)
	buffer.Write(cm.HMAC.HMAC)

	return buffer.Bytes(), nil
}

func ParseCapsuleMessage(packet *BasePacket, key []byte) (*CapsuleMessage, error) {
	decryptedPayload, err := DecryptPacket(packet.Payload, key, uint16(CipherTypeClassAES256CBC), uint16(len(key)))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt packet payload: %w", err)
	}

	if len(decryptedPayload) < BaseCapsuleMessageLength {
		return nil, fmt.Errorf("invalid packet length: %d", len(packet.Payload))
	}

	capsuleMessage := &CapsuleMessage{
		BaseHeader:    packet.BaseHeader,
		PayloadLength: binary.BigEndian.Uint16(decryptedPayload[0:2]),
		Padding:       binary.BigEndian.Uint16(decryptedPayload[2:4]),
	}

	capsuleMessage.Payload = decryptedPayload[4 : 4+int(capsuleMessage.PayloadLength)]

	// Validate the lengths of EndKey
	if len(capsuleMessage.Payload) != int(capsuleMessage.PayloadLength) {
		return nil, fmt.Errorf("invalid end key length: expected %d, got %d", capsuleMessage.PayloadLength, len(capsuleMessage.Payload))
	}

	// Check if the payload length matches the expected length
	expectedLength := BaseTunnelRequestLength + int(capsuleMessage.PayloadLength)
	// (BaseHeaderLength + HMACLength)

	if len(decryptedPayload) != expectedLength {
		return nil, fmt.Errorf("invalid packet length: expected %d, got %d", expectedLength, len(packet.Payload))
	}

	return capsuleMessage, nil
}
