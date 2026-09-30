package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/entity/service"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ repository.KeepAliveMonitor = (*keepAliveMonitor)(nil)

// TODO: CMSをフィールドとして持つべき
type keepAliveMonitor struct {
	connectionService   service.ConnectionService
	socketStore         repository.SocketStore
	baseHeader          entity.BaseHeader
	loginResponse       entity.LoginResponse
	localIPVersion      entity.TypeLocalIPVersion
	nsAddress           *entity.NSAddress
	keepAliveInterval   time.Duration
	keepAliveACKTimeout time.Duration
}

func NewKeepAliveMonitor(connectionService service.ConnectionService, socketStore repository.SocketStore, baseHeader entity.BaseHeader, loginResponse entity.LoginResponse, localIPVersion entity.TypeLocalIPVersion, nsAddress *entity.NSAddress, keepAliveIntervalSeconds int, keepAliveACKTimeoutSeconds int) repository.KeepAliveMonitor {
	return &keepAliveMonitor{
		connectionService:   connectionService,
		socketStore:         socketStore,
		baseHeader:          baseHeader,
		loginResponse:       loginResponse,
		localIPVersion:      localIPVersion,
		nsAddress:           nsAddress,
		keepAliveInterval:   time.Duration(keepAliveIntervalSeconds) * time.Second,
		keepAliveACKTimeout: time.Duration(keepAliveACKTimeoutSeconds) * time.Second,
	}
}

func (g *keepAliveMonitor) Run(ctx context.Context,
	keepAliveQueue chan<- entity.OutPacketQueue,
	keepAliveAckQueue <-chan struct{},
	reRegistrationQueue chan<- entity.OutPacketQueue,
) error {
	config.LogDebug("Starting keep alive sender")
	defer config.LogDebug("Finished keep alive sender")

	sendAddress, err := g.nsAddress.UDPAddr(g.localIPVersion)
	if err != nil {
		return fmt.Errorf("failed to determine local IP version: %w", err)
	}

	keepAlive := entity.GenerateKeepAlive(g.baseHeader)
	buffer, err := keepAlive.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal keep alive: %w", err)
	}

	t := time.NewTicker(g.keepAliveInterval)
	defer t.Stop()

	deadline := time.NewTimer(g.keepAliveACKTimeout)
	deadline.Stop()
	defer deadline.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			addr := sendAddress
			keepAliveQueue <- entity.OutPacketQueue{
				Buffer: buffer,
				Addr:   &addr,
			}

			deadline.Reset(g.keepAliveACKTimeout)
			select {
			case <-ctx.Done():
				deadline.Stop()
				return nil
			case <-keepAliveAckQueue:
				// TODO: 現状だと1周期遅延する可能性がある。今のところNS停止の検知を即座に行う必要はないため軽量な処理にしている。将来的に改善する必要が出てくるかもしれない。
				deadline.Stop()
				config.LogDebug("Keep-alive ACK received from NS")
			case <-deadline.C:
				config.LogDebug("Keep-alive ACK timed out")

				// Communication with CMS
				connectionResponse, err := g.connection(g.loginResponse)
				if err != nil {
					config.LogErr("failed to connection", "err", err)
					continue
				}

				// Update NS Address
				g.nsAddress.Set(connectionResponse.NSIPv4Address, connectionResponse.NSIPv6Address, connectionResponse.NSPort, connectionResponse.CommonKey)
				newAddr, err := g.nsAddress.UDPAddr(g.localIPVersion)
				if err != nil {
					config.LogErr("failed to determine local IP version", "err", err)
					continue
				}
				sendAddress = newAddr

				// Communication with NS
				buf, soc, err := g.generateRegistrationRequest()
				if err != nil {
					config.LogErr("failed to generate registration request", "err", err)
					continue
				}

				reRegistrationQueue <- entity.OutPacketQueue{
					Buffer: buf,
					Addr:   &newAddr,
					Soc:    soc,
				}
			}
		}
	}
}

func (g *keepAliveMonitor) connection(loginResponse entity.LoginResponse) (*entity.ConnectionResponse, error) {
	return g.connectionService.Connection(loginResponse)
}

func (g *keepAliveMonitor) generateRegistrationRequest() ([]byte, entity.Socket, error) {
	soc, err := g.socketStore.HighestPrioritySocket()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get highest priority socket: %w", err)
	}

	nodeIPv4Address := soc.LocalIPv4Addr()
	nodeIPv6Address := soc.LocalIPv6Addr()
	g.localIPVersion = entity.DetermineLocalIPVersion(nodeIPv4Address, nodeIPv6Address)
	g.socketStore.SwitchActiveConn(g.localIPVersion)

	registrationRequest := entity.GenerateRegistrationRequest(g.baseHeader, nodeIPv4Address, nodeIPv6Address, soc.InterfaceName())
	buffer, err := registrationRequest.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal re registration request: %w", err)
	}

	//TODO: keepAliveFlagの実装
	//g.keepAliveFlag = registrationResponse.KeepAliveFlag
	return buffer, soc, nil
}
