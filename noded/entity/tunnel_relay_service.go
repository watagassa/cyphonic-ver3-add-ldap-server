package entity

type TunnelRelayService struct {
	TunnelRelayServiceID   string `gorm:"column:trs_id"`
	TunnelRelayServiceIPv4 []byte `gorm:"column:trs_ipv4"`
	TunnelRelayServiceIPv6 []byte `gorm:"column:trs_ipv6"`
	TunnelRelayServicePort int    `gorm:"column:trs_port"`
	QUICFlag               bool   `gorm:"column:quic_flag"` // 0:UDP, 1:QUIC
}
