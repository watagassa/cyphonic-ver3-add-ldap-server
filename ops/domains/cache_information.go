// Package domains contains the database informations.
package domains

import "time"

// CacheInformations is to set Cache Informations.
// Ex. CYPHONIC Node's real IP addresses, NMS's real IP addresses.
type CacheInformations struct {
	Zone                string     `gorm:"column:zone"`
	NodeID              string     `gorm:"column:node_id"`
	RealIPv4            []byte     `gorm:"column:real_ipv4"`
	RealIPv6            []byte     `gorm:"column:real_ipv6"`
	CommonKey           []byte     `gorm:"column:common_key"`
	CommonKeyCipherType int        `gorm:"column:common_key_cipher_type"`
	CommonKeyLength     int        `gorm:"column:common_key_length"`
	CommonKeyExpire     time.Time  `gorm:"column:common_key_expire"`
	NMSFlag             bool       `gorm:"column:nms_flag"`
	CreatedAt           *time.Time `gorm:"column:created_at"`
	UpdatedAt           *time.Time `gorm:"column:updated_at"`
}
