package entity

import "time"

// NodeInformation is to set Node Information.
// Ex. CYPHONIC node's real IP addresses.
type NodeInformation struct {
	FQDN                string    `gorm:"column:fqdn"`
	NodeID              []byte    `gorm:"column:node_id"`
	VirtualIPv4         []byte    `gorm:"column:virtual_ipv4"`
	VirtualIPv6         []byte    `gorm:"column:virtual_ipv6"`
	ApplicationID       string    `gorm:"column:application_id"`
	ApplicationPort     int       `gorm:"column:application_port"`
	NotificationType    int       `gorm:"column:notification_type"`
	CommonKey           []byte    `gorm:"column:common_key"`
	CommonKeyCipherType int       `gorm:"column:common_key_cipher_type"`
	CommonKeyLength     int       `gorm:"column:common_key_length"`
	CommonKeyExpire     time.Time `gorm:"column:common_key_expire"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}
