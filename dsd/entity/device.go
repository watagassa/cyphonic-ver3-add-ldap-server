package entity

type Device struct {
	ID                         int    `gorm:"column:id;primaryKey"`
	DeviceID                   string `gorm:"column:device_id"`
	AccountID                  int    `gorm:"account_id"`
	DeviceName                 string `gorm:"column:device_name"`
	FQDN                       string `gorm:"column:fqdn"`
	DeviceTypeID               int    `gorm:"column:device_type_id"`
	AdapterFlag                bool   `gorm:"column:adapter_flag"`
	GeneralNodeFlag            bool   `gorm:"column:general_node_flag"`
	InterNodeCertificationFlag bool   `gorm:"column:inter_node_certification_flag"`
	GroupProfile               uint64 `gorm:"column:group_profile"`
	Status                     int    `gorm:"column:status"`
}
