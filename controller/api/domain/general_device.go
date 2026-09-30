// Package domain is the entity of the data model.
package domain

import (
	"time"
)

// GeneralDevice is to set a CYPHONIC general device.
type GeneralDevice struct {
	ID         int       `json:"id"`
	AdapterID  string    `json:"adapter_id"`
	DeviceID   string    `json:"device_id" validate:"required"`
	MacAddress string    `json:"mac_address"`
	EnableIPv6 bool      `json:"enable_ipv6"`
	AuthType   int       `json:"auth_type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
