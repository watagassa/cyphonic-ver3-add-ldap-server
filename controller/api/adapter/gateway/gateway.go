// Package gateway implements the entity of database operations.
package gateway

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/controller/api/domain"
	"github.com/Pluslab/cyphonic/controller/api/usecases/port"
	"gorm.io/gorm"
)

type Client struct {
	Conn *gorm.DB
}

type GormClientFactory interface {
	NewClient(ctx context.Context) (*Client, error)
}

type Gateway struct {
	Client *Client
}

func NewRepository(clientFactory GormClientFactory) (port.Repository, error) {
	client, err := clientFactory.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to return client: %w", err)
	}

	return &Gateway{
		Client: client,
	}, nil
}

func (gateway *Gateway) GetChildDeviceInformation(ctx context.Context, adapterID string) ([]domain.ChildDeviceInformation, error) {
	informations := make([]domain.ChildDeviceInformation, 0)

	/*
				// FYI: Query algorithm
				* MainTable: general_devices
				* JOIN-1: devices
				* JOIN-2: device_passwords
		        *
				* SELECT
				* general_devices.id,
				* devices.device_name,
				* device_passwords.device_id,
				* device_passwords.password,
				* general_devices.auth_type,
				* general_devices.mac_address,
				* general_devices.enable_ipv6,
				* general_devices.created_at,
				* general_devices.updated_at
				* FROM general_devices
				* INNER JOIN devices ON general_devices.device_id=devices.device_id
				* INNER JOIN device_passwords ON devices.device_id=device_passwords.device_id
				* WHERE general_devices.adapter_id=[adapter_id]
				* AND devices.general_node_flag=true
				* ORDER BY general_devices.id ASC;
	*/
	result := gateway.Client.Conn.Table("general_devices").Select(
		[]string{
			"general_devices.id", "devices.device_name",
			"device_passwords.device_id", "device_passwords.password", "general_devices.auth_type",
			"general_devices.mac_address", "general_devices.enable_ipv6",
			"general_devices.created_at", "general_devices.updated_at",
		}).Joins(
		"INNER JOIN devices ON general_devices.device_id=devices.device_id").Joins(
		"INNER JOIN device_passwords ON devices.device_id=device_passwords.device_id").Where(
		"general_devices.adapter_id = ? AND devices.general_node_flag = ?", adapterID, true).Order(
		"general_devices.id ASC").Scan(&informations)
	if result.Error != nil {
		return nil, result.Error
	}

	return informations, nil
}

func (gateway *Gateway) GetDeviceByDeviceID(ctx context.Context, deviceID string) (*domain.Device, error) {
	device := new(domain.Device)

	result := gateway.Client.Conn.Limit(1).Where("device_id = ?", deviceID).Find(&device)
	if result.Error != nil {
		return device, result.Error
	}

	return device, nil
}

func (gateway *Gateway) GetDevices(ctx context.Context) ([]domain.Device, error) {
	devices := make([]domain.Device, 0)

	result := gateway.Client.Conn.Find(&devices)
	if result.Error != nil {
		return nil, result.Error
	}

	return devices, nil
}

func (gateway *Gateway) GetGeneralDeviceByDeviceID(ctx context.Context, deviceID string) (*domain.GeneralDevice, error) {
	generalDevice := new(domain.GeneralDevice)

	result := gateway.Client.Conn.Limit(1).Where("device_id = ?", deviceID).Find(&generalDevice)
	if result.Error != nil {
		return generalDevice, result.Error
	}

	return generalDevice, nil
}

func (gateway *Gateway) GetGeneralDevices(ctx context.Context) ([]domain.GeneralDevice, error) {
	generalDevices := make([]domain.GeneralDevice, 0)

	result := gateway.Client.Conn.Find(&generalDevices)
	if result.Error != nil {
		return nil, result.Error
	}

	return generalDevices, nil
}

func (gateway *Gateway) GetAccountByAccountID(ctx context.Context, accountID int) (*domain.Account, error) {
	account := new(domain.Account)

	result := gateway.Client.Conn.Limit(1).Where("id = ?", accountID).Find(&account)
	if result.Error != nil {
		return account, result.Error
	}

	return account, nil
}

func (gateway *Gateway) CreateDevices(ctx context.Context, device *domain.Device) (*domain.Device, error) {
	newDevice := new(domain.Device)

	result := gateway.Client.Conn.Limit(1).Where("device_id = ? OR device_name = ? OR fqdn = ?", device.DeviceID, device.DeviceName, device.FQDN).Find(&newDevice)
	if result.Error == nil {
		if result.RowsAffected == 0 {
			gateway.Client.Conn.Create(&device)
		} else {
			return newDevice, fmt.Errorf("already exists")
		}
	} else {
		return newDevice, result.Error
	}

	return device, nil
}
