// Package domain is the entity of the data model.
package domain

import "time"

// Device is to set a CYPHONIC device.
type Device struct {
	ID                         int       `json:"id"`
	Version                    string    `json:"version"`
	DeviceID                   string    `json:"device_id" validate:"required"`
	AccountID                  int       `json:"account_id"`
	DeviceName                 string    `json:"device_name" validate:"required"`
	FQDN                       string    `json:"fqdn" validate:"required"`
	NodeManagementAreaID       int       `json:"node_management_area_id"`
	DeviceType                 int       `json:"device_type"`
	AdapterFlag                bool      `json:"adapter_flag"`
	GeneralNodeFlag            bool      `json:"general_node_flag"`
	InterNodeCertificationFlag bool      `json:"inter_node_certification_flag"`
	Status                     int       `json:"status"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}
