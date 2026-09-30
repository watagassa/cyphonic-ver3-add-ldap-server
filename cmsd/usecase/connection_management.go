//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/cmsd/$GOPACKAGE
package usecase

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/cmsd/entity"
	"github.com/Pluslab/cyphonic/cmsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/cmsd/usecase/repository"
	"golang.org/x/sync/errgroup"
)

const BufferSize = 4096

type ConnectionManagement interface {
	ConnectionManagement(ctx context.Context) error
}

type connectionManagement struct {
	packetHandlerRepository   repository.PacketHandler
	sqlHandlerRepository      repository.SQLHandler
	certficateStoreRepository repository.CertificateStore
	tokenVerifier             repository.TokenVerifier
	commonKeyGenerator        repository.CommonKeyGenerator
	tlsListenerRepository     repository.TLSListener
	quicListenerRepository    repository.QUICListener
}

func NewConnectionManagement(ph repository.PacketHandler, sh repository.SQLHandler, cs repository.CertificateStore, tv repository.TokenVerifier, cg repository.CommonKeyGenerator, tl repository.TLSListener, ql repository.QUICListener) ConnectionManagement {
	return &connectionManagement{
		packetHandlerRepository:   ph,
		sqlHandlerRepository:      sh,
		certficateStoreRepository: cs,
		tokenVerifier:             tv,
		commonKeyGenerator:        cg,
		tlsListenerRepository:     tl,
		quicListenerRepository:    ql,
	}
}

func (cm *connectionManagement) ConnectionManagement(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return cm.listenAndServeOverTLS(ctx, errGroup)
	})
	errGroup.Go(func() error {
		return cm.listenAndServeOverQUIC(ctx, errGroup)
	})

	config.LogInfo("Connection Management Service is running...")

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("error occurred: %s", err)
	}

	return nil
}

func (cm *connectionManagement) listenAndServeOverTLS(ctx context.Context, errGroup *errgroup.Group) error {
	config.LogInfo("Activating TLS server")

	for {
		conn, err := cm.tlsListenerRepository.Accept()
		if err != nil {
			return fmt.Errorf("accepting tls request error: %s", err)
		}

		config.LogDebug("TLS connection established", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())

		select {
		case <-ctx.Done():
			goto done
		default:
			errGroup.Go(func() error {
				err := cm.manage(ctx, conn)
				if err != nil {
					config.LogErr("[TLS] connection management failed", "error", err)
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

func (cm *connectionManagement) listenAndServeOverQUIC(ctx context.Context, errGroup *errgroup.Group) error {
	config.LogInfo("Activating QUIC server")

	for {
		conn, err := cm.quicListenerRepository.Accept(ctx)
		if err != nil {
			return fmt.Errorf("accepting quic request error: %w", err)
		}

		config.LogDebug("QUIC connection established", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())

		select {
		case <-ctx.Done():
			goto done
		default:
			errGroup.Go(func() error {
				err := cm.manage(ctx, conn)
				if err != nil {
					config.LogErr("[QUIC] connection management failed", "error", err)
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

func (cm *connectionManagement) manage(ctx context.Context, conn repository.ConnectionRepository) error {
	res := cm.packetHandlerRepository.GenerateConnectionResponse()

	defer func() {
		config.LogDebug("Generated Connection Response", "transaction_id", res.BaseHeader.TransactionID, "connection_response", res)

		b, err := res.ConvertBytes()
		if err != nil {
			config.LogErr("Failed to convert to bytes", "error", err)
		}

		if _, err := conn.Write(b); err != nil {
			config.LogErr("Failed to write", "error", err)
		} else {
			config.LogDebug("Sent Connection Response", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())
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
		cm.packetHandlerRepository.ChangeStatus(res, entity.StatusClassInternalServerError)

		return fmt.Errorf("read error: %w", err)
	}

	req, err := cm.packetHandlerRepository.Unmarshal(ctx, buf, length)
	if err != nil {
		cm.packetHandlerRepository.ChangeStatus(res, entity.StatusClassInternalServerError)

		return fmt.Errorf("connection request unmarshal error: %w", err)
	}

	config.LogDebug("Unmarshaled connection request", "transaction_id", req.BaseHeader.TransactionID, "connection_request", req)

	err = cm.tokenVerifier.VerifyAccessToken(req.AccessToken, req.BaseHeader.ID)
	if err != nil {
		cm.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassForbidden)

		return fmt.Errorf("verify access token error: %w", err)
	}

	config.LogDebug("Verified Access Token", "transaction_id", req.BaseHeader.TransactionID)

	nsInfo, status, err := cm.sqlHandlerRepository.GetRandomNotificationService()
	if err != nil {
		cm.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, status)

		return fmt.Errorf("get random notification error: %w", err)
	}

	config.LogDebug("Get random Notification Service", "ns_ipv4", nsInfo.IPv4, "ns_ipv6", nsInfo.IPv6, "ns_port", nsInfo.Port)

	commonKey, err := cm.commonKeyGenerator.GenerateCommonKey()
	if err != nil {
		cm.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError)

		return fmt.Errorf("get common key error: %w", err)
	}

	config.LogDebug("Generated CommonKey", "common_key", commonKey.CommonKey, "expire", commonKey.ExpireDate, "cipher_type", commonKey.CipherType)

	err = cm.sqlHandlerRepository.SetNodeInformation(*req, commonKey.CommonKey, commonKey.ExpireDate, commonKey.CipherType)
	if err != nil {
		cm.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError)

		return fmt.Errorf("set node information error: %w", err)
	}

	cm.packetHandlerRepository.SerializeNSInformation(res, nsInfo)
	cm.packetHandlerRepository.SerializeCommonKey(res, commonKey)
	cm.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassSuccess)

	return nil
}
