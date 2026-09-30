package entity

// VirtualIPAddress is to set a virtual IP address of CYPHONIC Node
type VirtualIPAddress struct {
	DeviceID           string `gorm:"column:device_id"`
	VirtualIPv4        []byte `gorm:"column:virtual_ipv4"`
	VirtualIPv4Netmask []byte `gorm:"column:virtual_ipv4_netmask"`
	VirtualIPv6        []byte `gorm:"column:virtual_ipv6"`
	VirtualIPv6Prefix  []byte `gorm:"column:virtual_ipv6_prefix"`
}
