package entity

// Device is to set a CYPHONIC device.
type Device struct {
	Version                    string `gorm:"column:version"`
	DeviceID                   string `gorm:"column:device_id"`
	AccountID                  int    `gorm:"account_id"`
	DeviceName                 string `gorm:"column:device_name"`
	FQDN                       string `gorm:"column:fqdn"`
	NodeManagementAreaID       int    `gorm:"column:node_management_area_id"`
	DeviceType                 int    `gorm:"column:device_type"`
	AdapterFlag                bool   `gorm:"column:adapter_flag"`
	GeneralNodeFlag            bool   `gorm:"column:general_node_flag"`
	InterNodeCertificationFlag bool   `gorm:"column:inter_node_certification_flag"`
	Status                     int    `gorm:"column:status"`
}

type DeviceTypeClass int

const (
	DeviceTypeClassLinux   DeviceTypeClass = 1
	DeviceTypeClassWindows DeviceTypeClass = 2
	DeviceTypeClassDarwin  DeviceTypeClass = 3
	DeviceTypeClassIOS     DeviceTypeClass = 4
	DeviceTypeClassAndroid DeviceTypeClass = 5
)
