// Package controller assembles InputPort, OutputPort, and Repository and executes InputPort.
package controller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Pluslab/cyphonic/controller/api/adapter/gateway"
	"github.com/Pluslab/cyphonic/controller/api/domain"
	"github.com/Pluslab/cyphonic/controller/api/usecases/port"
	"github.com/labstack/echo/v4"
)

// ListenAndServe is the first interface called when an API request is received.
type ListenAndServe interface {
	Healthz(ctx context.Context) func(c echo.Context) error
	GetChildDeviceInformations(ctx context.Context) func(c echo.Context) error
	GetDevice(ctx context.Context) func(c echo.Context) error
	GetDevices(ctx context.Context) func(c echo.Context) error
	GetGeneralDevice(ctx context.Context) func(c echo.Context) error
	GetGeneralDevices(ctx context.Context) func(c echo.Context) error
	GetAccounts(ctx context.Context) func(c echo.Context) error
	CreateDevices(ctx context.Context) func(c echo.Context) error
}

type OutputFactory func(echo.Context) port.OutputPort

type InputFactory func(port.OutputPort, port.Repository) port.InputPort

type RepositoryFactory func(gateway.GormClientFactory) (port.Repository, error)

// Controller assembles InputPort, OutputPort, and Repository.
type Controller struct {
	outputFactory     OutputFactory
	inputFactory      InputFactory
	repositoryFactory RepositoryFactory
	clientFactory     gateway.GormClientFactory
}

func NewController(outputFactory OutputFactory, inputFactory InputFactory, repositoryFactory RepositoryFactory, clientFactory gateway.GormClientFactory) ListenAndServe {
	return &Controller{
		outputFactory:     outputFactory,
		inputFactory:      inputFactory,
		repositoryFactory: repositoryFactory,
		clientFactory:     clientFactory,
	}
}

func (c *Controller) newInputPort(echo echo.Context) (port.InputPort, error) {
	outputPort := c.outputFactory(echo)

	repository, err := c.repositoryFactory(c.clientFactory)
	if err != nil {
		return nil, fmt.Errorf("failed to return repository: %w", err)
	}

	return c.inputFactory(outputPort, repository), nil
}

func (c *Controller) GetChildDeviceInformations(ctx context.Context) func(c echo.Context) error {
	return func(echo echo.Context) error {
		adapterID := echo.Param("adapter_id")

		inputPort, err := c.newInputPort(echo)
		if err != nil {
			return fmt.Errorf("failed to return child device information: %w", err)
		}

		return fmt.Errorf("%w", inputPort.GetChildDeviceInformation(ctx, adapterID))
	}
}

func (c *Controller) GetDevice(ctx context.Context) func(c echo.Context) error {
	return func(echo echo.Context) error {
		deviceID := echo.QueryParam("device_id")

		inputPort, err := c.newInputPort(echo)
		if err != nil {
			return fmt.Errorf("failed to return device by device id: %w", err)
		}

		return fmt.Errorf("%w", inputPort.GetDeviceByDeviceID(ctx, deviceID))
	}
}

func (c *Controller) GetDevices(ctx context.Context) func(c echo.Context) error {
	return func(echo echo.Context) error {
		inputPort, err := c.newInputPort(echo)
		if err != nil {
			return fmt.Errorf("failed to return get devices: %w", err)
		}

		return fmt.Errorf("%w", inputPort.GetDevices(ctx))
	}
}

func (c *Controller) GetGeneralDevice(ctx context.Context) func(c echo.Context) error {
	return func(echo echo.Context) error {
		deviceID := echo.QueryParam("device_id")

		inputPort, err := c.newInputPort(echo)
		if err != nil {
			return fmt.Errorf("failed to return get general devices by device id: %w", err)
		}

		return fmt.Errorf("%w", inputPort.GetGeneralDeviceByDeviceID(ctx, deviceID))
	}
}

func (c *Controller) GetGeneralDevices(ctx context.Context) func(c echo.Context) error {
	return func(echo echo.Context) error {
		inputPort, err := c.newInputPort(echo)
		if err != nil {
			return fmt.Errorf("failed to return get general devices: %w", err)
		}

		return fmt.Errorf("%w", inputPort.GetGeneralDevices(ctx))
	}
}

func (c *Controller) GetAccounts(ctx context.Context) func(c echo.Context) error {
	return func(echo echo.Context) error {
		accountID, err := strconv.Atoi(echo.QueryParam("account_id"))
		if err != nil {
			return fmt.Errorf("invalid request data: %w", err)
		}

		inputPort, err := c.newInputPort(echo)
		if err != nil {
			return fmt.Errorf("failed to return get account by account id: %w", err)
		}

		return fmt.Errorf("%w", inputPort.GetAccountByAccountID(ctx, accountID))
	}
}

func (c *Controller) CreateDevices(ctx context.Context) func(c echo.Context) error {
	return func(echo echo.Context) error {
		device := new(domain.Device)

		// TODO: validation for device_id, device_name, fqdn.
		if err := echo.Bind(&device); err != nil {
			return fmt.Errorf("invalid request data: %w", err)
		}

		inputPort, err := c.newInputPort(echo)
		if err != nil {
			return fmt.Errorf("failed to return create devices: %w", err)
		}

		return fmt.Errorf("%w", inputPort.CreateDevices(ctx, device))
	}
}
