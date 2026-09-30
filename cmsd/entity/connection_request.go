package entity

// ConnectionRequest is Connetion Request of a CYPHONIC packet.
type ConnectionRequest struct {
	BaseHeader        BaseHeader
	AccessTokenLength uint16
	Padding           uint16
	AccessToken       string
}

// Payload length in ConnectionRequest.
const (
	// AccessTokenLengthSize lengths (bytes).
	AccessTokenLengthSize = 2

	// PaddingSize lengths (bytes).
	PaddingSize = 2
)

// BaseConnectionRequestLength is length of (base) body part of Connection Request.
// Also, since the length of AccessToken is variable length, when
// marshaling, calculate and add.
const BaseConnectionRequestLength = BaseHeaderLen + AccessTokenLengthSize + PaddingSize
