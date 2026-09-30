package injector

import (
	"context"
	"errors"
	"fmt"
	"log"
	"syscall"

	"github.com/Pluslab/cyphonic/asd/controller"
	"github.com/Pluslab/cyphonic/asd/infrastructure"
	"github.com/Pluslab/cyphonic/asd/infrastructure/config"
	"github.com/Pluslab/cyphonic/asd/usecase"
	"github.com/Pluslab/cyphonic/asd/usecase/repository"
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

	config.LogInfo("Starting Authentication Service")

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

	idGenerator := i.IDGenerator()
	config.LogDebug("ID Generator initialization succeeded")

	tokenGenerator := i.TokenGenerator(cfg)
	config.LogDebug("Token Generator initialization succeeded")

	authentication := i.Authentication(packetHandler, sqlHandler, certificateStore, idGenerator, tokenGenerator, tlsListener, quicListener)
	controller := i.NewController(authentication)

	config.LogDebug("Authentication use case initialization succeeded")
	config.LogDebug("Controller initialization succeeded")

	if err := controller.Execute(ctx); err != nil {
		config.LogErr(fmt.Sprintf("failed to authenticate: %v", err))

		return exitStatusWithError
	}

	config.LogInfo("End Authentication Service")

	return exitStatusWithNoError
}

func (i *Injector) PacketHandler(cfg *config.Config) (repository.PacketHandler, error) {
	return infrastructure.NewPacketHandler(cfg)
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

func (i *Injector) IDGenerator() repository.IDGenerator {
	return infrastructure.NewIDGenerator()
}

func (i *Injector) TokenGenerator(cfg *config.Config) repository.TokenGenerator {
	return infrastructure.NewTokenGenerator(cfg)
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

func (i *Injector) Authentication(ph repository.PacketHandler, sh repository.SQLHandler, cs repository.CertificateStore, ig repository.IDGenerator, tg repository.TokenGenerator, tl repository.TLSListener, ql repository.QUICListener) usecase.Authentication {
	return usecase.NewAuthentication(ph, sh, cs, ig, tg, tl, ql)
}

func (i *Injector) NewController(authentication usecase.Authentication) controller.Controller {
	return controller.NewController(authentication)
}

func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY)
}
