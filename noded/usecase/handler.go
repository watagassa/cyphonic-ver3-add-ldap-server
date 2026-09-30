package usecase

import (
	"context"
	"errors"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
	"golang.org/x/sync/errgroup"
)

type Handler interface {
	Run(ctx context.Context) error

	ReceivePackets(ctx context.Context,
		receivedPacketQueue chan<- entity.InPacketQueue) error

	StageInboundMessages(ctx context.Context,
		receivedPacketQueue <-chan entity.InPacketQueue,
		parsedPacketQueue chan<- entity.ParsedPacketQueue,
		stagedInboundMsgQueue chan<- entity.StagedInboundMsgQueue,
		inboundMsgQueue chan<- *entity.InboundMsgQueue) error
	HandleInboundMessages(ctx context.Context,
		stagedInboundMsgQueue <-chan entity.StagedInboundMsgQueue) error
	HandlePackets(ctx context.Context,
		parsedPacketQueue <-chan entity.ParsedPacketQueue,
		generatedPacketQueue chan<- entity.OutPacketQueue,
		dnsAnswerQueue chan<- entity.DNSAnswerQueue,
		keepAliveAckQueue chan<- struct{}) error

	StageOutboundMessages(ctx context.Context,
		outRawMsgQueue <-chan entity.OutboundMsgQueue,
		stagedRawMsgQueue chan<- entity.StagedOutboundMsgQueue,
		capsuleMessageQueue chan<- *entity.CapsuleMsgQueue) error
	HandleOutboundMessages(ctx context.Context,
		stagedRawMsgQueue <-chan entity.StagedOutboundMsgQueue) error

	HandleDNSQueries(ctx context.Context,
		dnsQueue <-chan entity.DNSQueue,
		cachedDnsAnswerQueue chan<- entity.DNSAnswerQueue,
		directionRequestQueue chan<- entity.OutPacketQueue) error
	HandleIPAddrChanged(ctx context.Context,
		ipAddrChangedCh <-chan bool,
		reRegistrationQueue chan<- entity.OutPacketQueue) error

	SendPackets(ctx context.Context,
		capsuleMessageQueue <-chan *entity.CapsuleMsgQueue,
		generatedPacketQueue <-chan entity.OutPacketQueue,
		directionRequestQueue <-chan entity.OutPacketQueue,
		reRegistrationQueue <-chan entity.OutPacketQueue,
		keepAliveQueue <-chan entity.OutPacketQueue,
	) error
}

var _ Handler = (*handler)(nil)

const (
	maxPacketSize         = 4096
	numWorkers            = 4
	queueSize             = 65535
	keepAliveAckQueueSize = 1
)

var ErrUnknownPacketType = errors.New("unknown packet type")

type handler struct {
	tunHandler            repository.TunHandler
	tunDnsHandler         repository.TunDnsHandler
	netlinkListener       repository.NetlinkListener
	keepAliveMonitor      repository.KeepAliveMonitor
	routeDirectionCache   repository.RouteDirectionCache
	dnsCache              repository.DNSCache
	nsAddress             *entity.NSAddress
	nodeID                entity.ID
	socketStore           repository.SocketStore
	peerRegistry          repository.PeerRegistry
	baseHeader            entity.BaseHeader
	fqdn                  []byte
	localIPVersion        entity.TypeLocalIPVersion
	routeOptimizationMode bool
}

func NewHandler(
	tunHandler repository.TunHandler,
	tunDnsHandler repository.TunDnsHandler,
	netlinkListener repository.NetlinkListener,
	keepAliveMonitor repository.KeepAliveMonitor,
	routeDirectionCache repository.RouteDirectionCache,
	dnsCache repository.DNSCache,
	socketStore repository.SocketStore,
	peerRegistry repository.PeerRegistry,
	nsAddress *entity.NSAddress,
	nodeID entity.ID,
	baseHeader entity.BaseHeader,
	fqdn []byte,
	localIPVersion entity.TypeLocalIPVersion,
	routeOptimizationMode bool,
) Handler {
	return &handler{
		tunHandler:            tunHandler,
		tunDnsHandler:         tunDnsHandler,
		netlinkListener:       netlinkListener,
		keepAliveMonitor:      keepAliveMonitor,
		routeDirectionCache:   routeDirectionCache,
		dnsCache:              dnsCache,
		socketStore:           socketStore,
		peerRegistry:          peerRegistry,
		nsAddress:             nsAddress,
		nodeID:                nodeID,
		baseHeader:            baseHeader,
		fqdn:                  fqdn,
		localIPVersion:        localIPVersion,
		routeOptimizationMode: routeOptimizationMode,
	}
}

func (h *handler) Run(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)
	receivedPacketQueue := make(chan entity.InPacketQueue, queueSize)
	generatedPacketQueue := make(chan entity.OutPacketQueue, queueSize)
	directionRequestQueue := make(chan entity.OutPacketQueue, queueSize)
	reRegistrationQueue := make(chan entity.OutPacketQueue, queueSize)
	keepAliveQueue := make(chan entity.OutPacketQueue, queueSize)
	keepAliveAckQueue := make(chan struct{}, keepAliveAckQueueSize)
	dnsQueue := make(chan entity.DNSQueue, queueSize)
	dnsAnswerQueue := make(chan entity.DNSAnswerQueue, queueSize)
	cachedDnsAnswerQueue := make(chan entity.DNSAnswerQueue, queueSize)
	capsuleMessageQueue := make(chan *entity.CapsuleMsgQueue, queueSize)
	inboundMsgQueue := make(chan *entity.InboundMsgQueue, queueSize)
	outboundMsgQueue := make(chan entity.OutboundMsgQueue, queueSize)
	stagedOutboundMsgQueue := make(chan entity.StagedOutboundMsgQueue, queueSize)
	stagedInboundMsgQueue := make(chan entity.StagedInboundMsgQueue, queueSize)
	parsedPacketQueue := make(chan entity.ParsedPacketQueue, queueSize)
	ipAddrChangedCh := make(chan bool, 64)

	// cache
	errGroup.Go(func() error {
		return h.routeDirectionCache.Run(ctx)
	})
	errGroup.Go(func() error {
		return h.dnsCache.Run(ctx)
	})

	// generate keep alive
	errGroup.Go(func() error {
		return h.keepAliveMonitor.Run(ctx, keepAliveQueue, keepAliveAckQueue, reRegistrationQueue)
	})

	// I/O
	errGroup.Go(func() error {
		return h.ReceivePackets(ctx, receivedPacketQueue)
	})
	errGroup.Go(func() error {
		return h.SendPackets(ctx, capsuleMessageQueue, generatedPacketQueue, directionRequestQueue, reRegistrationQueue, keepAliveQueue)
	})

	// packet handling
	errGroup.Go(func() error {
		return h.HandlePackets(ctx, parsedPacketQueue, generatedPacketQueue, dnsAnswerQueue, keepAliveAckQueue)
	})

	// message handling
	errGroup.Go(func() error {
		return h.StageOutboundMessages(ctx, outboundMsgQueue, stagedOutboundMsgQueue, capsuleMessageQueue)
	})
	errGroup.Go(func() error {
		return h.HandleOutboundMessages(ctx, stagedOutboundMsgQueue)
	})
	errGroup.Go(func() error {
		return h.StageInboundMessages(ctx, receivedPacketQueue, parsedPacketQueue, stagedInboundMsgQueue, inboundMsgQueue)
	})
	errGroup.Go(func() error {
		return h.HandleInboundMessages(ctx, stagedInboundMsgQueue)
	})

	// dns handling
	errGroup.Go(func() error {
		return h.HandleDNSQueries(ctx, dnsQueue, cachedDnsAnswerQueue, directionRequestQueue)
	})

	// changed ipaddr handling
	errGroup.Go(func() error {
		return h.HandleIPAddrChanged(ctx, ipAddrChangedCh, reRegistrationQueue)
	})

	// tun handler
	errGroup.Go(func() error {
		return h.tunHandler.Read(ctx, outboundMsgQueue)
	})
	errGroup.Go(func() error {
		return h.tunHandler.Write(ctx, inboundMsgQueue)
	})

	// dns tun handler
	errGroup.Go(func() error {
		return h.tunDnsHandler.Read(ctx, dnsQueue)
	})
	errGroup.Go(func() error {
		return h.tunDnsHandler.Write(ctx, dnsAnswerQueue, cachedDnsAnswerQueue)
	})

	// listener
	errGroup.Go(func() error {
		return h.netlinkListener.WatchIPAddrChange(ctx, ipAddrChangedCh)
	})

	config.LogDebug("Node daemon is running...")

	return errGroup.Wait()
}
