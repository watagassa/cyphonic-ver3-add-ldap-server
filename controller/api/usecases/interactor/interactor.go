// Package interactor defines an InputPort entity that receives request data and a processing flow that leaves out the technical elements.
package interactor

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/controller/api/domain"
	"github.com/Pluslab/cyphonic/controller/api/usecases/port"
)

// Interactor hides technical elements by calling interfaces.
type Interactor struct {
	OutputPort port.OutputPort
	Repository port.Repository
}

func NewInputPort(outputPort port.OutputPort, repository port.Repository) port.InputPort {
	return &Interactor{
		OutputPort: outputPort,
		Repository: repository,
	}
}

func (i *Interactor) GetChildDeviceInformation(ctx context.Context, adapterID string) error {
	informations, err := i.Repository.GetChildDeviceInformation(ctx, adapterID)
	if err != nil {
		return fmt.Errorf("failed to return child device information: %w", i.OutputPort.OutputError(fmt.Errorf("failed to get child device information: %w", err)))
	}

	if len(informations) < 1 {
		return fmt.Errorf("failed to return child device information: %w", i.OutputPort.OutputSimpleMessage("no matching records found"))
	}

	return fmt.Errorf("%w", i.OutputPort.OutputGetChildDeviceInformation(informations))
}

func (i *Interactor) GetDeviceByDeviceID(ctx context.Context, deviceID string) error {
	device, err := i.Repository.GetDeviceByDeviceID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to return device information: %w", i.OutputPort.OutputError(fmt.Errorf("failed to get device information: %w", err)))
	}

	if device.ID < 1 {
		return fmt.Errorf("failed to return device information: %w", i.OutputPort.OutputSimpleMessage("no matching records found"))
	}

	return fmt.Errorf("%w", i.OutputPort.OutputGetDeviceByDeviceID(device))
}

func (i *Interactor) GetDevices(ctx context.Context) error {
	devices, err := i.Repository.GetDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to return device information: %w", i.OutputPort.OutputError(fmt.Errorf("failed to get device information: %w", err)))
	}

	if len(devices) < 1 {
		return fmt.Errorf("failed to return device information: %w", i.OutputPort.OutputSimpleMessage("no matching records found"))
	}

	return fmt.Errorf("%w", i.OutputPort.OutputGetDevices(devices))
}

func (i *Interactor) GetGeneralDeviceByDeviceID(ctx context.Context, deviceID string) error {
	generalDevice, err := i.Repository.GetGeneralDeviceByDeviceID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to return general device information: %w", i.OutputPort.OutputError(fmt.Errorf("failed to get general device information: %w", err)))
	}

	if generalDevice.ID < 1 {
		return fmt.Errorf("failed to return general device information: %w", i.OutputPort.OutputSimpleMessage("no matching records found"))
	}

	return fmt.Errorf("%w", i.OutputPort.OutputGetGeneralDeviceByDeviceID(generalDevice))
}

func (i *Interactor) GetGeneralDevices(ctx context.Context) error {
	generalDevices, err := i.Repository.GetGeneralDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to return general device information: %w", i.OutputPort.OutputError(fmt.Errorf("failed to get general device information: %w", err)))
	}

	if len(generalDevices) < 1 {
		return fmt.Errorf("failed to return general device information: %w", i.OutputPort.OutputSimpleMessage("no matching records found"))
	}

	return fmt.Errorf("%w", i.OutputPort.OutputGetGeneralDevices(generalDevices))
}

func (i *Interactor) GetAccountByAccountID(ctx context.Context, accountID int) error {
	account, err := i.Repository.GetAccountByAccountID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to return account information: %w", i.OutputPort.OutputError(fmt.Errorf("failed to get account information: %w", err)))
	}

	if account.ID < 1 {
		return fmt.Errorf("failed to return account information: %w", i.OutputPort.OutputSimpleMessage("no matching records found"))
	}

	return fmt.Errorf("%w", i.OutputPort.OutputGetAccountByAccountID(account))
}

func (i *Interactor) CreateDevices(ctx context.Context, device *domain.Device) error {
	device, err := i.Repository.CreateDevices(ctx, device)
	if err != nil {
		return fmt.Errorf("%w", i.OutputPort.OutputError(fmt.Errorf("failed to insert device information: %w", err)))
	}

	return fmt.Errorf("%w", i.OutputPort.OutputCreateDevices(device))
}
