package entity

import "time"

type NodeInformation struct {
	FQDN                    string     `gorm:"column:fqdn;"`
	NodeID                  []byte     `gorm:"column:node_id;"`
	DeviceID                []byte     `gorm:"column:device_id;"`
	DeviceTypeID            int        `gorm:"column:device_type_id;"`
	Version                 string     `gorm:"column:version;"`
	VirtualIPv4             []byte     `gorm:"column:virtual_ipv4;"`
	VirtualIPv4Netmask      []byte     `gorm:"column:virtual_ipv4_netmask;"`
	VirtualIPv6             []byte     `gorm:"column:virtual_ipv6;"`
	VirtualIPv6PrefixLength []byte     `gorm:"column:virtual_ipv6_prefix_length;"`
	ApplicationID           []byte     `gorm:"column:application_id;"`
	ApplicationPort         uint16     `gorm:"column:application_port;"`
	NotificationType        uint16     `gorm:"column:notification_type;"`
	GroupProfile            int        `gorm:"column:group_profile;"`
	CreatedAt               *time.Time `gorm:"column:created_at"`
	UpdatedAt               *time.Time `gorm:"column:updated_at"`
}
