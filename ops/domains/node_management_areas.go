// Package domains contains the database informations.
package domains

// NodeManagementArea is to set a CYPHONIC node management area.
type NodeManagementArea struct {
	ID   int    `gorm:"id"`
	Area string `gorm:"area"`
	FQDN string `gorm:"fqdn"`
}
