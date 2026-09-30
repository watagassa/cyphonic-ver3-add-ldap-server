package entity

// DevicePassword is to set a CYPHONIC device password.
type DevicePassword struct {
	DeviceID string `gorm:"column:device_id"`
	Password string `gorm:"column:password"`
}
