// Package domains contains the database informations.
package domains

// DevicePasswords is to set a CYPHONIC device password.
type DevicePasswords struct {
	DeviceID string `gorm:"column:device_id"`
	Password string `gorm:"password"`
}
