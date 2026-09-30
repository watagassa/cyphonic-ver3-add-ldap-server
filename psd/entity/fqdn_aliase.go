package entity

import "time"

// FQDNAlias is to set Node's Alias FQDN.
type FQDNAlias struct {
	DeviceID  string    `gorm:"column:device_id"`
	FQDN      string    `gorm:"column:fqdn_alias"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}
