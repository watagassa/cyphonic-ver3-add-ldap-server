// Package driver calls controller and sets technical elements.
package driver

import (
	"context"
	"net/http"

	"github.com/Pluslab/cyphonic/controller/api/adapter/controller"
	"github.com/Pluslab/cyphonic/controller/api/domain/validator"
	govalidator "github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ListenAndServe interface {
	Serve(ctx context.Context, addrPort string)
}

type Driver struct {
	echo       *echo.Echo
	controller controller.ListenAndServe
}

func NewDriver(echo *echo.Echo, controller controller.ListenAndServe) ListenAndServe {
	return &Driver{
		echo:       echo,
		controller: controller,
	}
}

func (driver *Driver) Serve(ctx context.Context, addrPort string) {
	driver.echo.HideBanner = true

	driver.echo.Validator = &validator.CustomValidator{
		Validator: govalidator.New(),
	}

	driver.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:          middleware.DefaultSkipper,
		AllowOrigins:     []string{"*"},
		AllowOriginFunc:  middleware.DefaultCORSConfig.AllowOriginFunc,
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowHeaders:     middleware.DefaultCORSConfig.ExposeHeaders,
		AllowCredentials: middleware.DefaultStaticConfig.IgnoreBase,
		ExposeHeaders:    middleware.DefaultCORSConfig.AllowHeaders,
		MaxAge:           middleware.DefaultCORSConfig.MaxAge,
	}))
	driver.echo.Use(middleware.Logger())
	driver.echo.Use(middleware.Recover())

	driver.echo.GET("/healthz", driver.controller.Healthz(ctx))

	api := driver.echo.Group("/api")
	api.GET("/device", driver.controller.GetDevice(ctx))
	api.GET("/devices", driver.controller.GetDevices(ctx))
	api.GET("/general_device", driver.controller.GetGeneralDevice(ctx))
	api.GET("/general_devices", driver.controller.GetGeneralDevices(ctx))
	api.GET("/accounts", driver.controller.GetAccounts(ctx))
	api.POST("/devices", driver.controller.CreateDevices(ctx))

	adapter := driver.echo.Group("/api/adapter")
	adapter.GET("/:adapter_id/children", driver.controller.GetChildDeviceInformations(ctx))

	driver.echo.Logger.Fatal(driver.echo.Start(addrPort))
}
