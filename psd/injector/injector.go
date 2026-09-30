package injector

import (
	"context"
	"errors"
	"fmt"
	"log"
	"syscall"

	"github.com/Pluslab/cyphonic/psd/controller"
	"github.com/Pluslab/cyphonic/psd/infrastructure"
	"github.com/Pluslab/cyphonic/psd/infrastructure/config"
	"github.com/Pluslab/cyphonic/psd/usecase"
	"github.com/Pluslab/cyphonic/psd/usecase/repository"
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

	config.LogInfo("Starting Provision Service")

	packetHandler, err := i.PacketHandler(cfg)
	if err != nil {
		config.LogErr(fmt.Sprintf("can't call Packet Handler: %v", err))

		return exitStatusWithError
	}

	config.LogDebug("Packet Handler initialization succeeded")

	sqlHandler, err := i.SQLHandler(cfg)
	if err != nil {
		config.LogErr(fmt.Sprintf("can't call SQL Handler: %v", err))

		return exitStatusWithError
	}

	config.LogDebug("SQL Handler initialization succeeded")

	certificateStore, err := i.CertificateStore(cfg)
	if err != nil {
		panic(fmt.Errorf("missing certificate: %w", err))
	}

	config.LogDebug("Certificate Store initialization succeeded")

	tokenVerifire := i.TokenVerifier(cfg)

	config.LogDebug("Token Verifire initialization succeeded")

	tlsListener, err := i.TLSListener(cfg, certificateStore)
	if err != nil {
		config.LogErr(fmt.Sprintf("can't call TLS Listener: %v", err))

		return exitStatusWithError
	}

	config.LogDebug("TLS Listener initialization succeeded")

	quicListener, err := i.QUICListener(cfg, certificateStore)
	if err != nil {
		config.LogErr(fmt.Sprintf("can't call QUIC Listener: %v", err))

		return exitStatusWithError
	}

	config.LogDebug("QUIC Listener initialization succeeded")

	provision := i.Provision(packetHandler, sqlHandler, certificateStore, tokenVerifire, tlsListener, quicListener)
	config.LogDebug("Provision use case initialization succeeded")

	controller := i.NewController(provision)
	config.LogDebug("Controller initialization succeeded")

	if err := controller.Execute(ctx); err != nil {
		config.LogErr(fmt.Sprintf("failed to provision: %v", err))

		return exitStatusWithError
	}

	config.LogInfo("End Provision Service")

	return exitStatusWithNoError
}

func (i *Injector) PacketHandler(cfg *config.Config) (repository.PacketHandler, error) {
	return infrastructure.NewPacketHandler(cfg.VIPVersion)
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

func (i *Injector) Provision(ph repository.PacketHandler, sh repository.SQLHandler, cs repository.CertificateStore, tv repository.TokenVerifier, tl repository.TLSListener, ql repository.QUICListener) usecase.Provision {
	return usecase.NewProvision(ph, sh, cs, tv, tl, ql)
}

func (i *Injector) NewController(provision usecase.Provision) controller.Controller {
	return controller.NewController(provision)
}

func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY)
}
