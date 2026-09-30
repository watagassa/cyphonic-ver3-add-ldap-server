// Package injector injects the dependencies and initializes the controller.
package injector

import (
	"context"
	"log"

	"github.com/Pluslab/cyphonic/controller/api/adapter/controller"
	"github.com/Pluslab/cyphonic/controller/api/adapter/gateway"
	"github.com/Pluslab/cyphonic/controller/api/adapter/presenter"
	"github.com/Pluslab/cyphonic/controller/api/driver"
	"github.com/Pluslab/cyphonic/controller/api/orm"
	"github.com/Pluslab/cyphonic/controller/api/pkg/config"
	"github.com/Pluslab/cyphonic/controller/api/pkg/logger"
	"github.com/Pluslab/cyphonic/controller/api/usecases/interactor"
	"github.com/labstack/echo/v4"
)

type Injector struct{}

func (i *Injector) Run(ctx context.Context) {
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}

	err = logger.InitZap(cfg.DebugMode)
	if err != nil {
		log.Fatal(err)
	}

	outputFactory := i.NewOutputFactory()
	inputFactory := i.NewInputFactory()
	repositoryFactory := i.NewRepositoryFactory()
	gormClientFactory := i.NewGormClientFactory(cfg)

	serve := i.NewController(outputFactory, inputFactory, repositoryFactory, gormClientFactory)
	driver := driver.NewDriver(echo.New(), serve)
	driver.Serve(ctx, ":"+cfg.Port)
}

// NewOutputFactory initializes OutputPort.
func (i *Injector) NewOutputFactory() controller.OutputFactory {
	return presenter.NewOutputPort
}

// NewInputFactory initializes InputPort.
func (i *Injector) NewInputFactory() controller.InputFactory {
	return interactor.NewInputPort
}

// NewRepositoryFactory initializes Repository.
func (i *Injector) NewRepositoryFactory() controller.RepositoryFactory {
	return gateway.NewRepository
}

// NewGormClientFactory initializes GormClient.
func (i *Injector) NewGormClientFactory(cfg *config.Config) gateway.GormClientFactory {
	return &orm.CustomGormClientFactory{Cfg: cfg}
}

// NewController initializes Controller.
func (i *Injector) NewController(opf controller.OutputFactory, ipf controller.InputFactory, rf controller.RepositoryFactory, cf gateway.GormClientFactory) controller.ListenAndServe {
	return controller.NewController(opf, ipf, rf, cf)
}
