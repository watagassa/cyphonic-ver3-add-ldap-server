package entity

// HolePunching is Hole Punching packet structure.
type HolePunching struct {
	BaseHeader BaseHeader
	HMAC       HMAC
}

// ParseHolePunching parse a structure from binary.
func ParseHolePunching(b *BasePacket) (HolePunching, error) {
	hp := HolePunching{
		BaseHeader: b.BaseHeader,
		HMAC:       b.Payload[:HMACLength],
	}

	return hp, nil
}
