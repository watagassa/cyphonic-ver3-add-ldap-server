package injector

import (
	"context"
	"errors"
	"fmt"
	"syscall"

	"github.com/Pluslab/cyphonic/dsd/controller"
	"github.com/Pluslab/cyphonic/dsd/infrastructure"
	"github.com/Pluslab/cyphonic/dsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/dsd/usecase"
	"github.com/Pluslab/cyphonic/dsd/usecase/repository"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

const (
	exitStatusWithNoError int = 0
	exitStatusWithError   int = 1
)

type Injector struct{}

func (i *Injector) NewInjector() *Injector {
	return &Injector{}
}

func (i *Injector) Run(ctx context.Context) int {
	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("failed to get cfg: %v", err)
		return exitStatusWithError
	}

	reset, err := config.InitZap(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to initiate logger: %w", err))
	}

	defer func() {
		if err := zap.S().Sync(); err != nil && !isIgnorableSyncError(err) {
			panic(fmt.Errorf("failed to flush log entries: %w", err))
		}

		reset()
	}()

	config.LogInfo("Starting Direction Service")

	// Create UDP connection
	udpConnRepository, err := i.NewUDPConn(cfg.Port)
	if err != nil {
		log.Fatalf("failed to create UDP connection: %v", err)
		return exitStatusWithError
	}

	// Create SQLHandler
	sqlHandlerRepository, err := i.NewSQLHandler(cfg.DatabaseHost, cfg.DatabaseUser, cfg.DatabaseName, cfg.DatabasePort, cfg.DatabaseSSLMode, cfg.DatabaseTimeZone, cfg.FQDN)
	if err != nil {
		log.Fatalf("failed to create SQL handler: %v", err)
		return exitStatusWithError
	}

	// Create Handler
	handler, err := usecase.NewHandler(udpConnRepository, sqlHandlerRepository)
	if err != nil {
		log.Fatalf("failed to create handler use case: %v", err)
		return exitStatusWithError
	}
	// Create Controller
	controller := controller.NewController(handler)

	config.LogDebug("Handler use case initialization succeeded")
	config.LogDebug("Controller initialization succeeded")

	if err := controller.Execute(ctx); err != nil {
		log.Fatalf("failed to execute controller: %v", err)
		return exitStatusWithError
	}

	return exitStatusWithNoError
}

func (i *Injector) NewUDPConn(port string) (repository.UDPConn, error) {
	udpConn, err := infrastructure.NewUDPConn(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %v", err)
	}

	return udpConn, nil
}

func (i *Injector) NewSQLHandler(host, user, name, port, sslMode, timeZone, fqdn string) (repository.SQLHandler, error) {
	sqlHandler, err := infrastructure.NewSQLHandler(host, user, name, port, sslMode, timeZone, fqdn)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL handler: %v", err)
	}

	return sqlHandler, nil
}

func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY)
}
