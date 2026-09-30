package entity

import "net/netip"

// NotificationService is to set a Notification Service.
type NotificationService struct {
	NSid     string `gorm:"column:ns_id"`
	NSIPv4   []byte `gorm:"column:ns_ipv4"`
	NSIPv6   []byte `gorm:"column:ns_ipv6"`
	NSPort   int    `gorm:"column:ns_port"`
	QuicFlag bool   `gorm:"column:quic_flag"`
}

type NotificationServiceInfomation struct {
	ID       string
	IPv4     netip.Addr
	IPv6     netip.Addr
	Port     uint16
	QuicFlag bool
}
