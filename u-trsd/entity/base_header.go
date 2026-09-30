package entity

import (
	"bytes"
	"encoding/binary"
)

// cyphonicVersion represents the version of CYPHONIC.
const cyphonicVersion = 3

// BaseHeader represents the header of a CYPHONIC packet.
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

// BaseHeaderLength defines the length of BaseHeader in bytes.
const BaseHeaderLength = 32

// ID represents a 16-byte CYPHONIC ID.
type ID []byte

// IDlen defines the length of ID in bytes.
const IDlen = 16

// TypeClass represents the type of a CYPHONIC packet.
type TypeClass uint8

// StatusClass represents the status of a CYPHONIC packet.
type StatusClass uint8

// Known TypeClass values.
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
	TypeClassNotificationConfirmation
	TypeClassFinalizationRequest
	TypeClassKeepAlive
)

// Known StatusClass values.
const (
	StatusClassSuccess StatusClass = iota
	StatusClassInternalServerError
	StatusClassFalsificationDetected
	StatusClassDecryptionFailed
	StatusClassAuthenticationFailed
	StatusClassForbidden
	StatusClassProvisionFailed
	StatusClassProvisionAliaseFQDNFailed
	StatusClassConnectionResolutionFailed
	StatusClassRegistrationFailed
	StatusClassDestinationNotFound
	StatusClassExecutionModeMismatch
	StatusClassProfileMismatch
	StatusClassTunnelConstructionFailed
	StatusClassFinalizationFailed
	StatusClassGeneralFailure
)

// Serialize converts the BaseHeader into a byte slice.
func (b *BaseHeader) Serialize() []byte {
	buf := make([]byte, 0, BaseHeaderLength)
	buffer := bytes.NewBuffer(buf)

	fields := []interface{}{
		b.TransactionID, b.Version, b.Type, b.Status,
		b.Count, b.SequenceNumber, b.MessageLength,
		b.Token, b.NextOpt,
	}

	for _, field := range fields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil
		}
	}

	buffer.Write(b.ID)
	return buffer.Bytes()
}

// serializeTransactionID is to serialize BaseHeader's TransactionID.
func (b *BaseHeader) serializeTransactionID(prevBaseHeader BaseHeader) {
	b.TransactionID = prevBaseHeader.TransactionID
}

// serializeVersion is to serialize BaseHeader's version.
func (b *BaseHeader) serializeVersion() {
	b.Version = cyphonicVersion
}

// serializeType is to serialize BaseHeader's type.
func (b *BaseHeader) serializeType(typ TypeClass) {
	b.Type = TypeClass(typ)
}

func (b *BaseHeader) serializeStatus(st StatusClass) {
	b.Status = StatusClass(st)
}

func (b *BaseHeader) serializeCount() {
	b.Count = 0
}

// serializeSequenceNumber is to serialize BaseHeader's SequenceNumber.
func (b *BaseHeader) serializeSequenceNumber(prevBaseHeader BaseHeader) {
	b.SequenceNumber = prevBaseHeader.SequenceNumber + 1
}

func (b *BaseHeader) serializeMessageLength(length uint16) {
	b.MessageLength = length
}

func (b *BaseHeader) serializeToken() {
	b.Token = 0
}

func (b *BaseHeader) serializeNextOpt() {
	b.NextOpt = 0
}

func (b *BaseHeader) serializeID(id []byte) {
	copy(b.ID, id)
}
