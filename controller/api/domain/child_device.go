// Package domain is the entity of the data model.
package domain

import "time"

// ChildDeviceInformation defines the format of general node information returned to CYPHONIC adapter.
// CYPHONIC adapter gets only general node (ChildDevice) information under management.
type ChildDeviceInformation struct {
	ID         int       `json:"id"`
	DeviceName string    `json:"device_name"`
	DeviceID   string    `json:"device_id"`
	Password   string    `json:"password"`
	AuthType   int       `json:"auth_type"`
	MacAddress string    `json:"mac_address"`
	EnableIPv6 bool      `json:"enable_ipv6"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
