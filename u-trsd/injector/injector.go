package injector

import (
	"context"
	"errors"
	"fmt"
	"log"
	"syscall"

	"github.com/Pluslab/cyphonic/u-trsd/controller"
	"github.com/Pluslab/cyphonic/u-trsd/infrastructure"
	"github.com/Pluslab/cyphonic/u-trsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/u-trsd/usecase"
	"github.com/Pluslab/cyphonic/u-trsd/usecase/repository"
	"go.uber.org/zap"
)

const (
	exitStatusWithNoError int = 0
	exitStatusWithError   int = 1
)

// Injector is a struct that initializes the application components.
type Injector struct{}

// Run initializes the application and runs the main logic.
func (i *Injector) Run(ctx context.Context) int {
	config.LogInfo("Getting config")
	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
		return exitStatusWithError
	}

	config.LogInfo("Initializing logger")
	reset, err := config.InitZap(cfg)
	if err != nil {
		log.Fatalf("failed to initiate logger: %v", err)
		return exitStatusWithError
	}

	config.LogInfo("Flushing log entries")
	defer func() {
		if err := zap.S().Sync(); err != nil && !isIgnorableSyncError(err) {
			panic(fmt.Errorf("failed to flush log entries: %w", err))
		}

		reset()
	}()

	config.LogInfo("Starting UDP Tunnel Relay Service")

	// Create UDP connection
	config.LogInfo("Creating UDP connection")
	udpConnRepository, err := i.newUDPConn(cfg.Port)
	if err != nil {
		log.Fatalf("failed to create UDP connection: %v", err)
		return exitStatusWithError
	}

	// Create SQLHandler
	sqlHandlerRepository, err := i.newSQLHandler(cfg.DatabaseHost, cfg.DatabaseUser, cfg.DatabaseName, cfg.DatabasePort, cfg.DatabaseSSLMode, cfg.DatabaseTimeZone, cfg.FQDN)
	if err != nil {
		log.Fatalf("failed to create SQL handler: %v", err)
		return exitStatusWithError
	}

	// Create Cache Handler
	pathCache, err := infrastructure.NewPathCache(sqlHandlerRepository)
	if err != nil {
		log.Fatalf("failed to create cache handler: %v", err)
		return exitStatusWithError
	}

	// Create Handler
	handler := usecase.NewHandler(udpConnRepository, sqlHandlerRepository, pathCache)
	// Create Controller
	controller := controller.NewController(handler)

	if err := controller.Execute(ctx); err != nil {
		log.Fatalf("failed to execute controller: %v", err)
		return exitStatusWithError
	}

	return exitStatusWithNoError
}

func (i *Injector) newUDPConn(port string) (repository.UDPConn, error) {
	udpConn, err := infrastructure.NewUDPConn(port)
	if err != nil {
		config.LogDebug("Unable to establish UDP connection", "error", err)
	}

	return udpConn, nil
}

func (i *Injector) newSQLHandler(host, user, name, port, sslMode, timeZone, fqdn string) (repository.SQLHandler, error) {
	sqlHandler, err := infrastructure.NewSQLHandler(host, user, name, port, sslMode, timeZone, fqdn)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL handler: %v", err)
	}

	return sqlHandler, nil
}

func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY)
}
