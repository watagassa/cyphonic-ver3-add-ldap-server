//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/noded/$GOPACKAGE
package daemon

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/entity/service"
	"github.com/Pluslab/cyphonic/noded/infrastructure"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
	"golang.org/x/sync/errgroup"
)

var _ Daemon = (*daemon)(nil)

type Daemon interface {
	Start(ctx context.Context) error
	Stop()
	Up() error
	Down() error
	GetStatus() string
	GetFqdn() string
	Provision(loginResponse entity.LoginResponse) error
	Connection(loginResponse entity.LoginResponse) error
	Registration(loginResponse entity.LoginResponse) error
	Finalization(loginResponse entity.LoginResponse) error
	LoginResponse() entity.LoginResponse
}

type daemon struct {
	isLogin                    bool
	readyCond                  *sync.Cond
	upc                        chan struct{}
	downc                      chan struct{}
	routeOptimizationMode      bool
	selfDesiredFQDN            []byte
	baseHeader                 entity.BaseHeader // TODO: baseHeaderを直接持たせるのは良くないかも
	loginResponse              entity.LoginResponse
	accessToken                []byte
	fqdn                       []byte
	localIPVersion             entity.TypeLocalIPVersion
	nodeID                     entity.ID
	l2Flag                     entity.L2FlagClass
	tunInterfaceName           string
	tunDnsInterfaceName        string
	tunDnsInterfaceIPv4        string
	tunDnsInterfaceIPv6        string
	virtualIPType              int
	virtualIPv4Address         netip.Addr
	virtualIPv6Address         netip.Addr
	virtualIPv4Prefix          uint16
	virtualIPv6Prefix          uint16
	nsAddress                  *entity.NSAddress
	keepAliveFlag              entity.KeepAliveFlagClass
	provisionService           service.ProvisionService
	connectionService          service.ConnectionService
	registrationService        service.RegistrationService
	finalizationService        service.FinalizationService
	socketStore                repository.SocketStore
	keepAliveIntervalSeconds   int
	keepAliveACKTimeoutSeconds int
}

func NewDaemon(
	selfDesiredFQDN []byte,
	tunInterfaceName string,
	tunDnsInterfaceName string,
	tunDnsInterfaceIPv4 string,
	tunDnsInterfaceIPv6 string,
	virtualIPType int,
	routeOptimizationMode bool,
	keepAliveIntervalSeconds int,
	keepAliveACKTimeoutSeconds int,
	provisionService service.ProvisionService,
	connectionService service.ConnectionService,
	registrationService service.RegistrationService,
	finalizationService service.FinalizationService,
	socketStore repository.SocketStore,
) Daemon {
	return &daemon{
		isLogin:               false,
		readyCond:             sync.NewCond(&sync.Mutex{}),
		upc:                   make(chan struct{}),
		downc:                 make(chan struct{}),
		routeOptimizationMode: routeOptimizationMode,
		selfDesiredFQDN:       selfDesiredFQDN,
		baseHeader:            entity.BaseHeader{},
		accessToken:           nil,
		fqdn:                  nil,
		tunInterfaceName:      tunInterfaceName,
		tunDnsInterfaceName:   tunDnsInterfaceName,
		tunDnsInterfaceIPv4:   tunDnsInterfaceIPv4,
		tunDnsInterfaceIPv6:   tunDnsInterfaceIPv6,
		virtualIPType:         virtualIPType,
		virtualIPv4Address:    netip.IPv4Unspecified(),
		virtualIPv6Address:    netip.IPv6Unspecified(),
		virtualIPv4Prefix:     0,
		virtualIPv6Prefix:     0,
		l2Flag:                entity.L2FlagClassInactive,
		nsAddress: &entity.NSAddress{
			IPv4Addr:  netip.Addr{},
			IPv6Addr:  netip.Addr{},
			Port:      0,
			CommonKey: nil,
		},
		keepAliveFlag:              entity.KeepAliveFlagClassInactive,
		provisionService:           provisionService,
		connectionService:          connectionService,
		registrationService:        registrationService,
		finalizationService:        finalizationService,
		socketStore:                socketStore,
		keepAliveIntervalSeconds:   keepAliveIntervalSeconds,
		keepAliveACKTimeoutSeconds: keepAliveACKTimeoutSeconds,
	}
}

func (d *daemon) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	errGroup, ctx := errgroup.WithContext(ctx)
	defer cancel()

	// initialize sockets deadline
	err := d.socketStore.ClearDeadlines()
	if err != nil {
		return fmt.Errorf("failed to clear socket deadlines: %w", err)
	}

	errGroup.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-d.upc:
			handler, err := d.setup()
			if err != nil {
				return fmt.Errorf("failed to setup: %w", err)
			}

			// launch Handler
			config.LogInfo("Starting handler")
			if err := handler.Run(ctx); err != nil {
				return fmt.Errorf("failed to run handler: %w", err)
			}
			return nil
		}
	})

	errGroup.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-d.downc:
			cancel()
			return nil
		}
	})

	config.LogDebug("Waiting for login...")
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	config.LogInfo("Node daemon has stopped.")

	return nil
}

func (d *daemon) setup() (usecase.Handler, error) {
	// cache
	dnsCache := infrastructure.NewDNSCache()
	routeDirectionCache := infrastructure.NewRouteDirectionCache()

	// registry
	peerRegistry := infrastructure.NewPeerRegistry()

	// monitor
	keepAliveMonitor := infrastructure.NewKeepAliveMonitor(d.connectionService, d.socketStore, d.baseHeader, d.loginResponse, d.localIPVersion, d.nsAddress, d.keepAliveIntervalSeconds, d.keepAliveACKTimeoutSeconds)

	// tun
	config.LogInfo("Setting up TUN interface", "interface_name", d.tunInterfaceName, "virtual_ipv4_address", d.virtualIPv4Address, "virtual_ipv6_address", d.virtualIPv6Address, "virtual_ipv4_prefix", d.virtualIPv4Prefix, "virtual_ipv6_prefix", d.virtualIPv6Prefix)
	tunHandler, err := infrastructure.NewTunHandler(d.virtualIPType, d.tunInterfaceName, d.virtualIPv4Address.String(), d.virtualIPv6Address.String(), int(d.virtualIPv4Prefix), int(d.virtualIPv6Prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to create tun handler: %w", err)
	}

	config.LogInfo("Setting up DNS_TUN interface", "interface_name", d.tunDnsInterfaceName, "tun_dns_interface_ipv4", d.tunDnsInterfaceIPv4, "tun_dns_interface_ipv6", d.tunDnsInterfaceIPv6)
	tunDnsHandler, err := infrastructure.NewTunDnsHandler(d.virtualIPType, d.tunDnsInterfaceName, d.tunDnsInterfaceIPv4, d.tunDnsInterfaceIPv6, d.baseHeader.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create tun handler: %w", err)
	}

	// listener
	config.LogInfo("Setting up netlink listener")
	netlinkListener, err := infrastructure.NewNetlinkListener()
	if err != nil {
		return nil, fmt.Errorf("failed to create netlink listener: %w", err)
	}

	// handler
	return usecase.NewHandler(tunHandler, tunDnsHandler, netlinkListener, keepAliveMonitor, routeDirectionCache, dnsCache, d.socketStore, peerRegistry, d.nsAddress, d.nodeID, d.baseHeader, d.fqdn, d.localIPVersion, d.routeOptimizationMode), nil
}

func (d *daemon) Stop() {
	d.isLogin = false
}

func (d *daemon) Up() error {
	select {
	case d.upc <- struct{}{}:
		config.LogInfo("Node daemon is starting...")
		return nil
	default:
		config.LogErr("failed to start node daemon")
		return fmt.Errorf("failed to start node daemon")
	}
}

func (d *daemon) Down() error {
	select {
	case d.downc <- struct{}{}:
		config.LogInfo("Node daemon is stopping...")
		return nil
	default:
		config.LogErr("failed to stop node daemon")
		return fmt.Errorf("failed to stop node daemon")
	}
}

func (d *daemon) GetStatus() string {
	if d.virtualIPv4Address.IsUnspecified() && d.virtualIPv6Address.IsUnspecified() {
		return "inactive"
	}

	return "active"
}

func (d *daemon) GetFqdn() string {
	if d.fqdn == nil {
		return ""
	}
	return string(d.fqdn)
}

func (d *daemon) Provision(loginResponse entity.LoginResponse) error {
	d.loginResponse = loginResponse
	provisionResponse, err := d.provisionService.Provision(loginResponse, d.selfDesiredFQDN)
	if err != nil {
		return fmt.Errorf("failed to provision: %w", err)
	}

	d.baseHeader = loginResponse.BaseHeader // TODO: baseHeaderを直接持たせるのは良くないかも
	d.fqdn = provisionResponse.FQDN
	d.virtualIPv4Address = provisionResponse.VirtualIPv4Address
	d.virtualIPv6Address = provisionResponse.VirtualIPv6Address
	d.virtualIPv4Prefix = provisionResponse.VirtualIPv4Prefix
	d.virtualIPv6Prefix = provisionResponse.VirtualIPv6Prefix
	d.l2Flag = provisionResponse.L2Flag
	d.nodeID = provisionResponse.BaseHeader.ID

	return nil
}

func (d *daemon) Connection(loginResponse entity.LoginResponse) error {
	connectionResponse, err := d.connectionService.Connection(loginResponse)
	if err != nil {
		return fmt.Errorf("failed to connection: %w", err)
	}

	d.nsAddress.Set(connectionResponse.NSIPv4Address,
		connectionResponse.NSIPv6Address,
		connectionResponse.NSPort,
		connectionResponse.CommonKey)

	return nil
}

func (d *daemon) Registration(loginResponse entity.LoginResponse) error {
	soc, err := d.socketStore.HighestPrioritySocket()
	if err != nil {
		return fmt.Errorf("failed to get highest priority socket: %w", err)
	}

	nodeIPv4Address := soc.LocalIPv4Addr()
	nodeIPv6Address := soc.LocalIPv6Addr()
	d.localIPVersion = entity.DetermineLocalIPVersion(nodeIPv4Address, nodeIPv6Address)
	d.socketStore.SwitchActiveConn(d.localIPVersion)

	registrationResponse, err := d.registrationService.Registration(loginResponse, nodeIPv4Address, nodeIPv6Address, d.nsAddress, soc.InterfaceName())
	if err != nil {
		return fmt.Errorf("failed to registration: %w", err)
	}

	d.keepAliveFlag = registrationResponse.KeepAliveFlag

	return nil
}

func (d *daemon) Finalization(loginResponse entity.LoginResponse) error {
	_, err := d.finalizationService.Finalization(loginResponse)
	if err != nil {
		return fmt.Errorf("failed to connection: %w", err)
	}

	d.virtualIPv4Address = netip.IPv4Unspecified()
	d.virtualIPv6Address = netip.IPv6Unspecified()
	d.virtualIPv4Prefix = 0
	d.virtualIPv6Prefix = 0
	d.fqdn = nil
	d.l2Flag = entity.L2FlagClassInactive
	d.nsAddress.Reset()
	d.keepAliveFlag = entity.KeepAliveFlagClassInactive

	return nil
}

func (d *daemon) LoginResponse() entity.LoginResponse {
	return d.loginResponse
}
