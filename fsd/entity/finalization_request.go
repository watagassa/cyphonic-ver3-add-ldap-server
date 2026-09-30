package entity

// FinalizationRequest is Finalization Request of a CYPHONIC packet.
type FinalizationRequest struct {
	BaseHeader BaseHeader
}

// BaseFinalizationRequestLength is length of (base) body part of FinalizationRequest.
// Also, since the length of FQDN is variable length, when
// marshaling, calculate and add.
const BaseFinalizationRequestLength = BaseHeaderLength
