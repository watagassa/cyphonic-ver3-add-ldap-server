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

// BaseHeaderLen defines the length of BaseHeader in bytes.
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
	TypeClassFinalizationRequest
	TypeClassKeepAlive
	TypeClassKeepAliveAck
	TypeClassMultiCastRequest
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

// Initialize sets up the BaseHeader with the given parameters.
func (b *BaseHeader) Initialize(packetType TypeClass, prev BaseHeader) {
	b.TransactionID = prev.TransactionID
	b.Version = cyphonicVersion
	b.Type = packetType
	b.Status = StatusClassSuccess
	b.Count = 0
	b.SequenceNumber = prev.SequenceNumber + 1
	b.MessageLength = prev.MessageLength
	b.Token = 0
	b.NextOpt = 0
	copy(b.ID, prev.ID)
}
