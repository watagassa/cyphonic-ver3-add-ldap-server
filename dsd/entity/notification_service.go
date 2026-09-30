package entity

type NotificationService struct {
	NSID     string `gorm:"column:ns_id"`
	NSIPv4   []byte `gorm:"column:ns_ipv4"`
	NSIPv6   []byte `gorm:"column:ns_ipv6"`
	NSPort   int    `gorm:"column:ns_port"`
	QUICFlag bool   `gorm:"column:quic_flag"`
}
