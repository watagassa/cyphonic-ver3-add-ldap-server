package injector

import (
	"context"
	"errors"
	"fmt"
	"log"
	"syscall"

	"github.com/Pluslab/cyphonic/cmsd/controller"
	"github.com/Pluslab/cyphonic/cmsd/infrastructure"
	"github.com/Pluslab/cyphonic/cmsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/cmsd/usecase"
	"github.com/Pluslab/cyphonic/cmsd/usecase/repository"
	"go.uber.org/zap"
)

const (
	exitStatusWithNoError = 0
	exitStatusWithError   = 1
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

	config.LogInfo("Starting Connection Management Service")

	packetHandler := i.PacketHandler()

	config.LogDebug("Packet Handler initialization succeeded")

	sqlHandler, err := i.SQLHandler(cfg)
	if err != nil {
		config.LogErr("Can't call SQL Handler", "error", err)

		return exitStatusWithError
	}

	config.LogDebug("SQL Handler initialization succeeded")

	certificateStore, err := i.CertificateStore(cfg)
	if err != nil {
		config.LogErr("Missing certificate", "error", err)

		return exitStatusWithError
	}

	config.LogDebug("Certificate Store initialization succeeded")

	tokenVerifirer := i.TokenVerifier(cfg)

	config.LogDebug("Token Verifire initialization succeeded")

	commonKeyGenerator := i.CommonKeyGenerator(cfg)

	config.LogDebug("Token Verifire initialization succeeded")

	tlsListener, err := i.TLSListener(cfg, certificateStore)
	if err != nil {
		config.LogErr("Can't call TLS Listener", "error", err)

		return exitStatusWithError
	}

	config.LogDebug("TLS Listener initialization succeeded")

	quicListener, err := i.QUICListener(cfg, certificateStore)
	if err != nil {
		config.LogErr("Can't call QUIC Listener", "error", err)

		return exitStatusWithError
	}

	config.LogDebug("QUIC Listener initialization succeeded")

	connection_manager := i.ConnectionManagement(packetHandler, sqlHandler, certificateStore, tokenVerifirer, commonKeyGenerator, tlsListener, quicListener)
	config.LogDebug("Connection Management use case initialization succeeded")

	controller := i.NewController(connection_manager)
	config.LogDebug("Controller initialization succeeded")

	if err := controller.Execute(ctx); err != nil {
		config.LogErr("Failed to connection management", "error", err)

		return exitStatusWithError
	}

	config.LogInfo("End Connection Management Service")

	return exitStatusWithNoError
}

func (i *Injector) PacketHandler() repository.PacketHandler {
	return infrastructure.NewPacketHandler()
}

func (i *Injector) SQLHandler(cfg *config.Config) (repository.SQLHandler, error) {
	sqlHandler, err := infrastructure.NewSQLHandler(cfg.DatabaseHost, cfg.DatabaseUser,
		cfg.DatabaseName, cfg.DatabasePort, cfg.DatabaseSSLMode, cfg.DatabaseTimeZone, cfg.FQDN)
	if err != nil {
		return sqlHandler, fmt.Errorf("can't create SQL Handler: %w", err)
	}

	return sqlHandler, nil
}

func (i *Injector) CertificateStore(cfg *config.Config) (repository.CertificateStore, error) {
	certficateStore, err := infrastructure.NewCertificateStore(cfg.FQDN, cfg.RootCertName, cfg.AccessKey, cfg.SecretKey, cfg.Region, cfg.EndPoint, cfg.BucketName, cfg.CertificateLocation)
	if err != nil {
		return certficateStore, fmt.Errorf("can't create Certficate Store: %w", err)
	}

	return certficateStore, nil
}

func (i *Injector) TokenVerifier(cfg *config.Config) repository.TokenVerifier {
	return infrastructure.NewTokenVerifire(cfg)
}

func (i *Injector) CommonKeyGenerator(cfg *config.Config) repository.CommonKeyGenerator {
	return infrastructure.NewCommonKeyGenerator(cfg)
}

func (i *Injector) TLSListener(cfg *config.Config, certStore repository.CertificateStore) (repository.TLSListener, error) {
	tlsListener, err := infrastructure.NewTLSListener(cfg, certStore)
	if err != nil {
		return tlsListener, fmt.Errorf("can't create TLS Listener: %w", err)
	}

	return tlsListener, nil
}

func (i *Injector) QUICListener(cfg *config.Config, certStore repository.CertificateStore) (repository.QUICListener, error) {
	quicListener, err := infrastructure.NewQUICListener(cfg, certStore)
	if err != nil {
		return quicListener, fmt.Errorf("can't create QUIC Listener: %w", err)
	}

	return quicListener, nil
}

func (i *Injector) ConnectionManagement(ph repository.PacketHandler, sh repository.SQLHandler, cs repository.CertificateStore, tv repository.TokenVerifier, cg repository.CommonKeyGenerator, tl repository.TLSListener, ql repository.QUICListener) usecase.ConnectionManagement {
	return usecase.NewConnectionManagement(ph, sh, cs, tv, cg, tl, ql)
}

func (i *Injector) NewController(connectionManagement usecase.ConnectionManagement) controller.Controller {
	return controller.NewController(connectionManagement)
}

func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY)
}
