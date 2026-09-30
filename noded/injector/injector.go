package injector

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"

	"github.com/Pluslab/cyphonic/noded/application"
	"github.com/Pluslab/cyphonic/noded/application/daemon"
	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure"
	"github.com/Pluslab/cyphonic/noded/infrastructure/authenticationservice"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/infrastructure/connectionservice"
	"github.com/Pluslab/cyphonic/noded/infrastructure/finalizationservice"
	"github.com/Pluslab/cyphonic/noded/infrastructure/provisioningservice"
	"github.com/Pluslab/cyphonic/noded/infrastructure/registrationservice"
	"github.com/Pluslab/cyphonic/noded/sysconfig"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

var ErrFailedToAppendCACert = errors.New("failed to append ca certificate")

type Injector struct{}

func NewInjector() *Injector {
	return &Injector{}
}

func (i *Injector) Run(ctx context.Context) error {
	config.LogInfo("Getting config")
	cfg, err := i.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	config.LogInfo("Initializing logger")
	reset, err := config.InitZap(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to initiate logger: %w", err))
	}

	config.LogInfo("Flushing log entries")
	defer func() {
		if err := zap.S().Sync(); err != nil {
			panic(fmt.Errorf("failed to flush log entries: %w", err))
		}

		reset()
	}()

	// Rewrite DNS Resolv
	switch runtime.GOOS {
	case "linux", "darwin":
		sysconfig.RewriteResolv(cfg.VirtualIPType, cfg.TunDnsInterfaceIPv4, cfg.TunDnsInterfaceIPv6)
	}

	command := cobra.Command{
		Use:   "cyphonic",
		Short: "CYPHONIC provides secure and continuous e2e communication",
		Long:  "CYPHONIC provides secure and continuous e2e communication",
	}

	runCommand := i.setupRunCommand(ctx, cfg)
	command.AddCommand(runCommand)

	config.LogInfo("Starting Node daemon")
	err = command.ExecuteContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

func (i *Injector) setupRunCommand(ctx context.Context, cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "CYPHONIC runs the daemon and the API server",
		Long:  "CYPHONIC runs the daemon and the API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Provisioning Serviceとの通信用の証明書に変更したものも作成する
			config.LogInfo("Creating TLS config")
			tlsConfig, err := i.TLSConfig(cfg.AuthenticationService.CaFile)
			if err != nil {
				return fmt.Errorf("failed to create tls config: %w", err)
			}

			config.LogInfo("Creating socket store repository")
			socketStoreRepository, err := i.SocketStoreRepository(cfg.NodePort)
			if err != nil {
				return fmt.Errorf("failed to create socket store: %w", err)
			}

			// get one socket for the registration process
			config.LogDebug("Getting highest priority socket for registration")
			socket, err := socketStoreRepository.HighestPrioritySocket()
			if err != nil {
				return fmt.Errorf("failed to get highest priority socket: %w", err)
			}

			config.LogDebug("Use highest priority socket",
				"interface_name", socket.InterfaceName(),
				"local_addr_v4", socket.LocalIPv4String(),
				"local_addr_v6", socket.LocalIPv6String())

			authRepository := i.AuthRepository(net.JoinHostPort(cfg.AuthenticationService.Host, cfg.AuthenticationService.Port), tlsConfig)
			provisionRepository := i.ProvisionRepository(tlsConfig)
			connectionRepository := i.ConnectionRepository(tlsConfig)
			registrationRepository := i.RegistrationRepository(socket) // TODO: 適切なSocketを使用する
			finalizationRepository := i.FinalizationRepository(tlsConfig)

			errGroup, ctx := errgroup.WithContext(ctx)
			daemon := daemon.NewDaemon(
				cfg.SelfDesiredFQDN, cfg.TunInterfaceName, cfg.TunDnsInterfaceName,
				cfg.TunDnsInterfaceIPv4, cfg.TunDnsInterfaceIPv6, cfg.VirtualIPType, cfg.RouteOptimizationMode,
				cfg.KeepAliveIntervalSeconds, cfg.KeepAliveACKTimeoutSeconds,
				provisionRepository, connectionRepository, registrationRepository, finalizationRepository,
				socketStoreRepository,
			)

			// Start Daemon
			errGroup.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					default:
						if err := daemon.Start(ctx); err != nil {
							config.LogErr("Failed to run handler", "error", err)
							return fmt.Errorf("failed to start daemon, retrying: %w", err)
						}
					}
				}
			})

			// Start API Server
			errGroup.Go(func() error {
				if err := i.StartAPIServer(ctx, authRepository, daemon, cfg.APIPort); err != nil {
					return fmt.Errorf("failed to start API server: %w", err)
				}

				return nil
			})

			// Set up signal capturing
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, unix.SIGINT, unix.SIGTERM)

			// Wait for signals
			<-sigChan

			// TODO: 仮のシャットダウン処理なので要改善
			err = daemon.Finalization(daemon.LoginResponse())
			if err != nil {
				config.LogErr("Failed to finalize daemon", "error", err)
			}
			config.LogInfo("Daemon finalized")

			slog.Info("Shutting down...")

			return nil
		},
	}
}

func (i *Injector) Config() (*config.Config, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return cfg, nil
}

func (i *Injector) TLS(host, port, caFile string) (*tls.Conn, error) {
	CAPool := x509.NewCertPool()

	if caCert, err := os.ReadFile(caFile); err != nil {
		return nil, fmt.Errorf("failed to read ca certificate: %w", err)
	} else {
		if ok := CAPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("%w", ErrFailedToAppendCACert)
		}
	}

	config := &tls.Config{
		RootCAs: CAPool,
	}

	authenticationService := net.JoinHostPort(host, port)

	conn, err := tls.Dial("tcp", authenticationService, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return conn, nil
}

func (i *Injector) TLSConfig(caFile string) (*tls.Config, error) {
	CAPool := x509.NewCertPool()

	if caCert, err := os.ReadFile(caFile); err != nil {
		return nil, fmt.Errorf("failed to read ca certificate: %w", err)
	} else {
		if ok := CAPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("%w", ErrFailedToAppendCACert)
		}
	}

	config := &tls.Config{
		RootCAs: CAPool,
	}

	return config, nil
}

func (i *Injector) NewUDPConn(port string) (repository.UDPConn, error) {
	udpConn, err := infrastructure.NewUDPConn(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	return udpConn, nil
}

func (i *Injector) SocketStoreRepository(port string) (repository.SocketStore, error) {
	socketStore, err := infrastructure.NewSocketStore(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket store: %w", err)
	}

	return socketStore, nil
}

func (i *Injector) AuthRepository(authenticationServiceAddress string, tlsConfig *tls.Config) *authenticationservice.AuthenticationServiceRepository {
	return authenticationservice.NewAuthenticationServiceRepository(authenticationServiceAddress, tlsConfig)
}

func (i *Injector) ProvisionRepository(tlsConfig *tls.Config) *provisioningservice.ProvisionServiceRepository {
	return provisioningservice.NewProvisionServiceRepository(tlsConfig)
}

func (i *Injector) ConnectionRepository(tlsConfig *tls.Config) *connectionservice.ConnectionServiceRepository {
	return connectionservice.NewConnectionServiceRepository(tlsConfig)
}

func (i *Injector) RegistrationRepository(socket entity.Socket) *registrationservice.RegistrationServiceRepository {
	return registrationservice.NewRegistrationServiceRepository(socket)
}

func (i *Injector) FinalizationRepository(tlsConfig *tls.Config) *finalizationservice.FinalizationServiceRepository {
	return finalizationservice.NewFinalizationServiceRepository(tlsConfig)
}

func (i *Injector) StartAPIServer(ctx context.Context, authRepository repository.AuthRepository, daemon daemon.Daemon, apiPort string) error {
	// Start API Server
	server := application.NewServer(authRepository, daemon)
	engine := gin.New()
	application.RegisterHandlers(engine, server)
	engine.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", apiPort),
		Handler: engine,
	}

	err := application.Run(ctx, httpServer)
	if err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}

	return nil
}
