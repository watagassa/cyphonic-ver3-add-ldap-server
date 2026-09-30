package injector

import (
	"context"
	"errors"
	"fmt"
	"log"
	"syscall"

	"github.com/Pluslab/cyphonic/u-nsd/controller"
	"github.com/Pluslab/cyphonic/u-nsd/infrastructure"
	"github.com/Pluslab/cyphonic/u-nsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/u-nsd/usecase"
	"github.com/Pluslab/cyphonic/u-nsd/usecase/repository"

	"go.uber.org/zap"
)

const (
	exitStatusWithNoError int = 0
	exitStatusWithError   int = 1
)

type Injector struct{}

func (i *Injector) Run(ctx context.Context) int {
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get config: %w", err))
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

	config.LogInfo("Starting UDP Notification Service")

	udpConn, err := i.NewUDPConn(cfg.Port, cfg.DSIPv4, cfg.DSIPv6, cfg.DSPort)
	if err != nil {
		config.LogErr(fmt.Sprintf("can't call UDP Conn: %v", err))

		return exitStatusWithError
	}

	sqlHandler, err := i.NewSQLHandler(
		cfg.DatabaseHost,
		cfg.DatabaseUser,
		cfg.DatabaseName,
		cfg.DatabasePort,
		cfg.DatabaseSSLMode,
		cfg.DatabaseTimeZone,
		cfg.FQDN,
		cfg.NSIPv4,
		cfg.NSIPv6,
		cfg.Port,
	)
	if err != nil {
		config.LogErr(fmt.Sprintf("failed to create SQL Handler: %v", err))

		return exitStatusWithError
	}

	handler := usecase.NewHandler(udpConn, sqlHandler)
	controller := controller.NewController(handler)

	config.LogDebug("Handler use case initialization succeeded")
	config.LogDebug("Controller initialization succeeded")

	if err := controller.Execute(ctx); err != nil {
		config.LogErr(fmt.Sprintf("failed to notice: %v", err))

		return exitStatusWithError
	}

	config.LogInfo("End UDP Notification Service")

	return exitStatusWithNoError
}

func (i *Injector) NewUDPConn(port, dsIPv4, dsIPv6, dsPort string) (repository.UDPConn, error) {
	udpConn, err := infrastructure.NewUDPConn(port, dsIPv4, dsIPv6, dsPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %v", err)
	}

	return udpConn, nil
}

func (i *Injector) NewSQLHandler(host, username, databaseName, databasePort, sslMode, timezone, fqdn, nsIPv4, nsIPv6, strPort string) (repository.SQLHandler, error) {
	dbConn, err := infrastructure.NewSQLHandler(host, username, databaseName, databasePort, sslMode, timezone, fqdn, nsIPv4, nsIPv6, strPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create new SQLHandler: %w", err)
	}

	return dbConn, nil
}

func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY)
}
