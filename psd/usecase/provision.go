//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/psd/$GOPACKAGE
package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/Pluslab/cyphonic/psd/entity"
	"github.com/Pluslab/cyphonic/psd/infrastructure"
	"github.com/Pluslab/cyphonic/psd/infrastructure/config"
	"github.com/Pluslab/cyphonic/psd/usecase/repository"
	"golang.org/x/sync/errgroup"
)

const BufferSize = 4096

type Provision interface {
	Provision(ctx context.Context) error
}

type provision struct {
	packetHandlerRepository   repository.PacketHandler
	sqlHandlerRepository      repository.SQLHandler
	certficateStoreRepository repository.CertificateStore
	TokenVerifier             repository.TokenVerifier
	tlsListenerRepository     repository.TLSListener
	quicListenerRepository    repository.QUICListener
}

func NewProvision(ph repository.PacketHandler, sh repository.SQLHandler, cs repository.CertificateStore, tv repository.TokenVerifier, tl repository.TLSListener, ql repository.QUICListener) Provision {
	return &provision{
		packetHandlerRepository:   ph,
		sqlHandlerRepository:      sh,
		certficateStoreRepository: cs,
		TokenVerifier:             tv,
		tlsListenerRepository:     tl,
		quicListenerRepository:    ql,
	}
}

func (pr *provision) Provision(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return pr.listenAndServeOverTLS(ctx, errGroup)
	})
	errGroup.Go(func() error {
		return pr.listenAndServeOverQUIC(ctx, errGroup)
	})

	config.LogInfo("Provision service is running...")

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("error occurred: %s", err)
	}

	return nil
}

func (pr *provision) listenAndServeOverTLS(ctx context.Context, errGroup *errgroup.Group) error {
	config.LogInfo("Activated TLS server")

	for {
		conn, err := pr.tlsListenerRepository.Accept()
		if err != nil {
			return fmt.Errorf("accepting tls request error: %s", err)
		}

		config.LogDebug("TLS connection established", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())

		select {
		case <-ctx.Done():
			goto done
		default:
			errGroup.Go(func() error {
				err := pr.provisioning(ctx, conn)
				if err != nil {
					config.LogErr("[TLS] provisioning failed", "error", err)
				}
				return nil
			})
		}
	}

done:
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("errGroup detected error in TLS server: %w", err)
	}

	return nil
}

func (pr *provision) listenAndServeOverQUIC(ctx context.Context, errGroup *errgroup.Group) error {
	config.LogInfo("Activated QUIC server")

	for {
		conn, err := pr.quicListenerRepository.Accept(ctx)
		if err != nil {
			return fmt.Errorf("accepting quic request error: %w", err)
		}

		config.LogDebug("QUIC connection established", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())

		select {
		case <-ctx.Done():
			goto done
		default:
			errGroup.Go(func() error {
				err := pr.provisioning(ctx, conn)
				if err != nil {
					config.LogErr("[QUIC] provisioning failed: %w", err)
				}
				return nil
			})
		}
	}

done:
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("errGroup detected error in QUIC server: %w", err)
	}

	return nil
}

func (pr *provision) provisioning(ctx context.Context, conn repository.ConnectionRepository) error {
	var aliasFQDN entity.FQDNAlias

	res := pr.packetHandlerRepository.GenerateProvisionResponse()
	status := entity.StatusClassSuccess

	defer func() {
		config.LogDebug("Generated provision Response", "transaction_id", res.BaseHeader.TransactionID, "provision_response", res)

		b, err := res.ConvertBytes()
		if err != nil {
			config.LogErr("Failed to convert to bytes", "error", err)
		}

		if _, err := conn.Write(b); err != nil {
			config.LogErr("Failed to write", "error", err)
		} else {
			config.LogDebug("Sent Provision Response", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())
		}

		if err := conn.Close(); err != nil {
			config.LogErr("Error occurred when closing connection", "error", err)
		} else {
			config.LogDebug("Connection closed", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())
		}
	}()

	buf := make([]byte, BufferSize)

	length, err := conn.Read(buf)
	if err != nil {
		pr.packetHandlerRepository.ChangeStatus(res, entity.StatusClassInternalServerError)

		return fmt.Errorf("read error: %w", err)
	}

	req, err := pr.packetHandlerRepository.Unmarshal(ctx, buf, length)
	if err != nil {
		pr.packetHandlerRepository.ChangeStatus(res, entity.StatusClassInternalServerError)

		return fmt.Errorf("login request unmarshal error: %w", err)
	}

	config.LogDebug("Unmarshaled provision request", "transaction_id", req.BaseHeader.TransactionID, "provision_request", req)

	err = pr.TokenVerifier.VerifyAccessToken(req.AccessToken, req.BaseHeader.ID)
	if err != nil {
		pr.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassForbidden)

		return fmt.Errorf("verify access token error: %w", err)
	}

	nodeInfo, err := pr.sqlHandlerRepository.GetNodeInformation(ctx, req.BaseHeader.ID)
	if err != nil {
		pr.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassProvisionFailed)

		return fmt.Errorf("get node information error: %w", err)
	}

	virtualIPv4, virtualIPv6, err := convertAddressFromBytes(nodeInfo.VirtualIPv4, nodeInfo.VirtualIPv4Netmask, nodeInfo.VirtualIPv6, nodeInfo.VirtualIPv6Prefix)
	if err != nil {
		pr.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError)

		return fmt.Errorf("failed to marshal virtual ip addresses: %w", err)
	}

	if req.DesiredFQDNLength == 0 {
		goto serialize
	}

	config.LogDebug("Serching alias FQDN", "transaction_id", req.BaseHeader.TransactionID, "desired_fqdn", req.DesiredFQDN)

	aliasFQDN, err = pr.sqlHandlerRepository.GetAliasFQDN(ctx, nodeInfo.DeviceID, req.DesiredFQDN)
	if err == infrastructure.ErrAliasFQDNNotFound {
		config.LogWarn("Can't find alias FQDN", "transaction_id", req.BaseHeader.TransactionID, "device_id", nodeInfo.DeviceID, "desired_fqdn", string(req.DesiredFQDN), "error", err)
		status = entity.StatusClassProvisionAliaseFQDNFailed
		goto serialize
	} else if err != nil {
		pr.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError)

		return fmt.Errorf("failed to update fqdn in database: %w", err)
	}

	nodeInfo.FQDN = aliasFQDN.FQDN
	if err := pr.sqlHandlerRepository.UpdateNodeInformation(ctx, nodeInfo); err != nil {
		pr.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError)

		return fmt.Errorf("failed to update fqdn in database: %w", err)
	}

serialize:
	pr.packetHandlerRepository.SerializeVirtualIPAddress(res, virtualIPv4, virtualIPv6)
	pr.packetHandlerRepository.SerializeL2Flag(res)
	pr.packetHandlerRepository.SerializeFQDN(res, nodeInfo.FQDN)
	pr.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, status)

	return nil
}

func convertAddressFromBytes(ipv4Address, ipv4Netmask, ipv6Address, ipv6Prefix []byte) (ipv4 netip.Prefix, ipv6 netip.Prefix, err error) {
	if len(ipv4Address) != 4 {
		return ipv4, ipv6, errors.New("invalid length of ipv4 address")
	}

	if len(ipv6Address) != 16 {
		return ipv4, ipv6, errors.New("invalid length of ipv6 address")
	}

	var ipv4Slice [4]byte
	var ipv6Slice [16]byte

	copy(ipv4Slice[:], ipv4Address)
	copy(ipv6Slice[:], ipv6Address)
	ipv4Addr := netip.AddrFrom4(ipv4Slice)
	ipv6Addr := netip.AddrFrom16(ipv6Slice)

	ipv4PrefixLength := calcPrefixLength(ipv4Netmask)
	if !(1 <= ipv4PrefixLength && ipv4PrefixLength <= 30) {
		return ipv4, ipv6, errors.New("invalid prefix for virtual ipv4 address")
	}

	ipv6PrefixLength := calcPrefixLength(ipv6Prefix)
	if !(1 <= ipv6PrefixLength && ipv6PrefixLength <= 126) {
		return ipv4, ipv6, errors.New("invalid prefix for virtual ipv6 address")
	}

	ipv4 = netip.PrefixFrom(ipv4Addr, ipv4PrefixLength)
	ipv6 = netip.PrefixFrom(ipv6Addr, ipv6PrefixLength)

	return ipv4, ipv6, nil
}

func calcPrefixLength(mask []byte) int {
	length := 0
	findedZero := false

	for _, b := range mask {
		for i := 7; i >= 0; i-- {
			if b&(1<<i) != 0 {
				if findedZero {
					return -1
				} else {
					length++
				}
			} else {
				findedZero = true
			}
		}
	}

	return length
}
