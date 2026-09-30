package entity

// HMACLength lengths (bytes).
const HMACLength = 16

// HMAC is Hash-based Message Authentication Code of a CYPHONIC packet.
// length of the byte slice: a 16-byte slice.
type HMAC []byte

// GenerateHMAC is to generate HMAC.
func GenerateHMAC() (HMAC, error) {
	// TODO: Implement HMAC generation logic.
	hmacMock := make(HMAC, HMACLength)
	for i := 0; i < HMACLength; i++ {
		hmacMock[i] = 0x00
	}
	return hmacMock, nil
}

// DecodeHMAC is to decode HMAC to determine integrity.
func DecodeHMAC(hm HMAC, packet *BasePacket) (bool, error) {
	// TODO: Implement HMAC decoding logic.
	return true, nil
}

// GetHMACField extract the HMAC field from the CYPHONIC packet.
func GetHMACField(packet *BasePacket) (HMAC, error) {
	if len(packet.Payload) < HMACLength {
		return nil, ErrInvalidPacket
	}

	hmac := make(HMAC, HMACLength)
	copy(hmac, packet.Payload[len(packet.Payload)-HMACLength:])

	return hmac, nil
}
