package entity

// TunnelRelayService means Tunnel Relay Service information stored in the database.
type TunnelRelayService struct {
	TunnleRelayServiceID   string `gorm:"column:trs_id"`
	TunnleRelayServiceIPv4 []byte `gorm:"column:trs_ipv4"`
	TunnleRelayServiceIPv6 []byte `gorm:"column:trs_ipv6"`
	TunnleRelayServicePort int    `gorm:"column:trs_port"`
	QUICFlag               bool   `gorm:"column:quic_flag"` // 0:UDP, 1:QUIC
}
