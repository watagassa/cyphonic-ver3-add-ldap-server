// Package domains contains the database informations.
package domains

import "time"

// DigitalCertificates is to set a CYPHONIC device.
type DigitalCertificates struct {
	ID       int       `gorm:"column:id"`
	DeviceID string    `gorm:"column:device_id"`
	Status   int       `gorm:"column:status"`
	Expire   time.Time `gorm:"column:expire"`
}
