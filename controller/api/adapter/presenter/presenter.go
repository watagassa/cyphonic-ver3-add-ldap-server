// Package presenter implements the entity of OutputGetPort with technical elements.
package presenter

import (
	"fmt"
	"net/http"

	"github.com/Pluslab/cyphonic/controller/api/domain"
	"github.com/Pluslab/cyphonic/controller/api/domain/message"
	"github.com/Pluslab/cyphonic/controller/api/usecases/port"
	"github.com/labstack/echo/v4"
)

type Presenter struct {
	ctx echo.Context
}

func NewOutputPort(ctx echo.Context) port.OutputPort {
	return &Presenter{
		ctx: ctx,
	}
}

// OutputError returns an error.
func (p *Presenter) OutputError(err error) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusInternalServerError, message.CustomError{ErrorMessage: err.Error()}))
}

// OutputSimpleMessage returns simple message.
func (p *Presenter) OutputSimpleMessage(msg string) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, message.SimpleResponse{Message: msg}))
}

func (p *Presenter) OutputGetChildDeviceInformation(informations []domain.ChildDeviceInformation) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, informations))
}

func (p *Presenter) OutputGetDeviceByDeviceID(device *domain.Device) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, device))
}

func (p *Presenter) OutputGetDevices(devices []domain.Device) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, devices))
}

func (p *Presenter) OutputGetGeneralDeviceByDeviceID(generalDevice *domain.GeneralDevice) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, generalDevice))
}

func (p *Presenter) OutputGetGeneralDevices(generalDevices []domain.GeneralDevice) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, generalDevices))
}

func (p *Presenter) OutputGetAccountByAccountID(account *domain.Account) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, account))
}

func (p *Presenter) OutputCreateDevices(device *domain.Device) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, device))
}
