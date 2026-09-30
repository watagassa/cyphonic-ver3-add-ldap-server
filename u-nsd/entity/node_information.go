package entity

import "time"

// NodeInformation is to set Node Information.
// Ex. CYPHONIC node's real IP addresses.
type NodeInformation struct {
	Version                 string    `gorm:"column:version"`
	FQDN                    string    `gorm:"column:fqdn"`
	DeviceID                string    `gorm:"column:device_id"`
	NodeID                  []byte    `gorm:"column:node_id"`
	VirtualIPv4             []byte    `gorm:"column:virtual_ipv4"`
	VirtualIPv4Netmask      []byte    `gorm:"column:virtual_ipv4_netmask"`
	VirtualIPv6             []byte    `gorm:"column:virtual_ipv6"`
	VirtualIPv6PrefixLength []byte    `gorm:"column:virtual_ipv6_prefix_length"`
	ApplicationID           string    `gorm:"column:application_id"`
	ApplicationPort         int       `gorm:"column:application_port"`
	NotificationType        int       `gorm:"column:notification_type"`
	DeviceTypeID            int       `gorm:"column:device_type_id"`
	GroupProfile            uint64    `gorm:"column:group_profile"`
	CreatedAt               time.Time `gorm:"column:created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at"`
}
