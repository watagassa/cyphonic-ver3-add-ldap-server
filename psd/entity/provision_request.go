package entity

// ProvisionRequest is Provision Request of a CYPHONIC packet.
type ProvisionRequest struct {
	BaseHeader        BaseHeader
	DesiredFQDNLength uint16
	AccessTokenLength uint16
	DesiredFQDN       string
	AccessToken       string
}

// Payload length in ProvisionRequest.
const (
	// DesiredFQDNLengthSize lengths (bytes).
	DesiredFQDNLengthSize = 2

	// AccessTokenLengthSize lengths (bytes).
	AccessTokenLengthSize = 2
)

// BaseProvisionRequestLength is length of (base) body part of ProvisionRequest.
// Also, since the length of FQDN is variable length, when
// marshaling, calculate and add.
const BaseProvisionRequestLength = BaseHeaderLen + DesiredFQDNLengthSize + AccessTokenLengthSize
