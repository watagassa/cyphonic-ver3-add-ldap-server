package entity

import (
	"time"
)

type NodeAddress struct {
	NodeID        []byte     `gorm:"column:node_id"`
	InterfaceName string     `gorm:"column:interface_name"`
	RealIPv4      []byte     `gorm:"column:real_ipv4"`
	RealIPv6      []byte     `gorm:"column:real_ipv6"`
	NATIPv4       []byte     `gorm:"column:nat_ipv4"`
	NATIPv6       []byte     `gorm:"column:nat_ipv6"`
	NATPort       int        `gorm:"column:nat_port"`
	NSID          string     `gorm:"column:ns_id"`
	CreatedAt     *time.Time `gorm:"column:created_at;gorm:<-:create"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;gorm:<-:update"`
}
