package entity

// HMACLen lengths (bytes).
const HMACLength = 16

type HMAC struct {
	// HMAC is Hash-based Message Authentication Code of a CYPHONIC packet.
	HMAC []byte
}
