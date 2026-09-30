// Package port defines a basic interface that accepts requests and returns responses.
// Requests obtained from InputPort are processed by Interector and retrieved from OutPutPort.
// InputPort(interface) -> Interactor(entity) -> OutputPort(interface)
package port

import (
	"context"

	"github.com/Pluslab/cyphonic/controller/api/domain"
)

// InputPort gets request data.
type InputPort interface {
	Healthz(ctx context.Context) error
	GetChildDeviceInformation(ctx context.Context, adapterID string) error
	GetDeviceByDeviceID(ctx context.Context, deviceID string) error
	GetDevices(ctx context.Context) error
	GetGeneralDeviceByDeviceID(ctx context.Context, deviceID string) error
	GetGeneralDevices(ctx context.Context) error
	GetAccountByAccountID(ctx context.Context, accountID int) error
	CreateDevices(ctx context.Context, device *domain.Device) error
}

// OutputPort returns response data or an error.
type OutputPort interface {
	OutputSimpleMessage(string) error
	OutputError(error) error
	OutputGetChildDeviceInformation([]domain.ChildDeviceInformation) error
	OutputGetDeviceByDeviceID(*domain.Device) error
	OutputGetDevices([]domain.Device) error
	OutputGetGeneralDeviceByDeviceID(*domain.GeneralDevice) error
	OutputGetGeneralDevices([]domain.GeneralDevice) error
	OutputGetAccountByAccountID(*domain.Account) error
	OutputCreateDevices(*domain.Device) error
}

// Repository return the basic interface for database operations.
type Repository interface {
	GetChildDeviceInformation(ctx context.Context, adapterID string) ([]domain.ChildDeviceInformation, error)
	GetDeviceByDeviceID(ctx context.Context, deviceID string) (*domain.Device, error)
	GetDevices(ctx context.Context) ([]domain.Device, error)
	GetGeneralDeviceByDeviceID(ctx context.Context, deviceID string) (*domain.GeneralDevice, error)
	GetGeneralDevices(ctx context.Context) ([]domain.GeneralDevice, error)
	GetAccountByAccountID(ctx context.Context, accountID int) (*domain.Account, error)
	CreateDevices(ctx context.Context, device *domain.Device) (*domain.Device, error)
}
