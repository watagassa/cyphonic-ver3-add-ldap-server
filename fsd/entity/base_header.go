package entity

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// cyphonicVersion is mesuare version of CYPHONIC.
const cyphonicVersion = 3

// BaseHeader is the header of a CYPHONIC packet.
// CYPHONIC expects to send multiple signaling messages at one time.
// Example: RegistrationRequest and NotificationRequest send at one time.
// Count the number of signaling messages and add to the BaseHeader struct's Count.
type BaseHeader struct {
	TransactionID  uint32
	Version        uint8
	Type           TypeClass
	Status         StatusClass
	Count          uint8
	SequenceNumber uint32
	MessageLength  uint16
	Token          uint8
	NextOpt        uint8
	ID             ID
}

// BaseHeaderLen lengths (bytes).
const BaseHeaderLength = 32

// TransactionID lengths (bytes).
const TransactionIDLength = 4

// SequenceNumber lengths (bytes).
const SequenceNumberLength = 4

// ID is a CYPHONIC ID, a slice of bytes.
// length of the byte slice: a 16-byte slice.
type ID []byte

// IDlen lengths (bytes).
const IDlength = 16

// TypeClass defines the class associated with a CYPHONIC packet type.
// classes can be thought of as an array of parallel namespace trees.
type TypeClass uint8

// TypeClass known values.
const (
	TypeClassDefault TypeClass = iota
	TypeClassAck
	TypeClassNack
	TypeClassHolePunching
	TypeClassLoginRequest
	TypeClassLoginResponse
	TypeClassProvisionRequest
	TypeClassProvisionResponse
	TypeClassConnectionRequest
	TypeClassConnectionResponse
	TypeClassRegistrationRequest
	TypeClassRegistrationResponse
	TypeClassDirectionRequest
	TypeClassRouteDirectionToResponder
	TypeClassRouteDirectionConfirmation
	TypeClassRouteDirectionToInitiator
	TypeClassTunnelRequestForUDP
	TypeClassTunnelResponseForUDP
	TypeClassTunnelRequestForQUIC
	TypeClassTunnelResponseForQUIC
	TypeClassCapsuleMessageFromInitiator
	TypeClassCapsuleMessageFromResponder
	TypeClassHolePunchingForOptimization
	TypeClassAckForOptimization
	TypeClassFinalizationRequest
	TypeClassKeepAlive
	TypeClassKeepAliveAck
	TypeClassMultiCastRequest
)

// StatusClass defines the class associated with a CYPHONIC packet status.
// classes can be thought of as an array of parallel namespace trees.
type StatusClass uint8

// StatusClass known values.
const (
	StatusClassSuccess                    StatusClass = 0
	StatusClassInternalServerError        StatusClass = 1
	StatusClassFalsificationDetected      StatusClass = 2
	StatusClassDecryptionFailed           StatusClass = 3
	StatusClassAuthenticationFailed       StatusClass = 4
	StatusClassForbidden                  StatusClass = 5
	StatusClassProvisionFailed            StatusClass = 6
	StatusClassProvisionAliaseFQDNFailed  StatusClass = 7
	StatusClassConnectionResolutionFailed StatusClass = 8
	StatusClassRegistrationFailed         StatusClass = 9
	StatusClassDestinationNotFound        StatusClass = 10
	StatusClassExecutionModeMismatch      StatusClass = 11
	StatusClassProfileMismatch            StatusClass = 12
	StatusClassTunnelConstructionFailed   StatusClass = 13
	StatusClassFinalizationFailed         StatusClass = 14
	StatusClassGeneralFailure             StatusClass = 15
)

const MAXUint32 = 4294967295

func (b *BaseHeader) All() []byte {
	buf := make([]byte, 0, BaseHeaderLength)
	buffer := bytes.NewBuffer(buf)

	headerFields := []interface{}{
		b.TransactionID, b.Version, b.Type, b.Status,
		b.Count, b.SequenceNumber, b.MessageLength,
		b.Token, b.NextOpt,
	}

	for _, field := range headerFields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil
		}
	}

	buffer.Write(b.ID)

	return buffer.Bytes()
}

func (b *BaseHeader) SerializeBaseHeader(packetType TypeClass, statusClass StatusClass, messageLength uint16) {
	b.serializeType(packetType)
	b.serializeStatus(statusClass)
	b.serializeSequenceNumber()
	b.serializeMessageLength(messageLength)
}

func (b *BaseHeader) serializeType(packetType TypeClass) {
	b.Type = packetType
}

func (b *BaseHeader) serializeStatus(statusClass StatusClass) {
	b.Status = statusClass
}

func (b *BaseHeader) serializeSequenceNumber() {
	b.SequenceNumber++
}

func (b *BaseHeader) serializeMessageLength(messageLength uint16) {
	b.MessageLength = messageLength
}

// GenerateBaseHeader is to generate BaseHeader.
func GenerateBaseHeader(packetType TypeClass) (*BaseHeader, error) {
	transactionID, err := randomUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random uint32: %w", err)
	}

	// TODO: Implement set ID.
	return &BaseHeader{
		TransactionID:  transactionID,
		Version:        cyphonicVersion,
		Type:           packetType,
		Status:         StatusClassSuccess,
		Count:          0,
		SequenceNumber: 0,
		MessageLength:  0,
		Token:          0,
		NextOpt:        0,
		ID:             make([]byte, IDlength),
	}, nil
}

// randomUint32 is to generate a random uint32.
func randomUint32() (uint32, error) {
	var random uint32

	err := binary.Read(rand.Reader, binary.BigEndian, &random)
	if err != nil {
		return 0, fmt.Errorf("failed to read random: %w", err)
	}

	return random, nil
}

// UnmarshalBaseHeader is to unmarshal BaseHeader.
func UnmarshalBaseHeader(data []byte) *BaseHeader {
	return &BaseHeader{
		TransactionID:  binary.BigEndian.Uint32(data[0:4]),
		Version:        data[4],
		Type:           TypeClass(data[5]),
		Status:         StatusClass(data[6]),
		Count:          data[7],
		SequenceNumber: binary.BigEndian.Uint32(data[8:12]),
		MessageLength:  binary.BigEndian.Uint16(data[12:14]),
		Token:          data[14],
		NextOpt:        data[15],
		ID:             data[16:32],
	}
}
