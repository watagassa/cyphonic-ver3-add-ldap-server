//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/asd/$GOPACKAGE
package usecase

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/asd/entity"
	"github.com/Pluslab/cyphonic/asd/infrastructure/config"
	"github.com/Pluslab/cyphonic/asd/usecase/repository"
	"golang.org/x/sync/errgroup"
)

const BufferSize = 4096

type Authentication interface {
	Authenticate(ctx context.Context) error
}

type authentication struct {
	packetHandlerRepository   repository.PacketHandler
	sqlHandlerRepository      repository.SQLHandler
	certficateStoreRepository repository.CertificateStore
	idGenerator               repository.IDGenerator
	tokenGenerator            repository.TokenGenerator
	tlsListenerRepository     repository.TLSListener
	quicListenerRepository    repository.QUICListener
}

func NewAuthentication(ph repository.PacketHandler, sh repository.SQLHandler, cs repository.CertificateStore, ig repository.IDGenerator, tg repository.TokenGenerator, tl repository.TLSListener, ql repository.QUICListener) Authentication {
	return &authentication{
		packetHandlerRepository:   ph,
		sqlHandlerRepository:      sh,
		certficateStoreRepository: cs,
		idGenerator:               ig,
		tokenGenerator:            tg,
		tlsListenerRepository:     tl,
		quicListenerRepository:    ql,
	}
}

func (au *authentication) Authenticate(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return au.listenAndServeOverTLS(ctx, errGroup)
	})
	errGroup.Go(func() error {
		return au.listenAndServeOverQUIC(ctx, errGroup)
	})

	config.LogInfo("Authentication service is running...")

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("error occurred: %s", err)
	}

	return nil
}

func (au *authentication) listenAndServeOverTLS(ctx context.Context, errGroup *errgroup.Group) error {
	config.LogInfo("Activated TLS server")

	for {
		conn, err := au.tlsListenerRepository.Accept()
		if err != nil {
			return fmt.Errorf("accepting tls request error: %s", err)
		}

		config.LogDebug("TLS connection established", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())

		select {
		case <-ctx.Done():
			goto done
		default:
			errGroup.Go(func() error {
				err := au.auth(ctx, conn, false)
				if err != nil {
					config.LogErr("[TLS] authentication failed", "error", err)
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

func (au *authentication) listenAndServeOverQUIC(ctx context.Context, errGroup *errgroup.Group) error {
	config.LogInfo("Activated QUIC server")

	for {
		conn, err := au.quicListenerRepository.Accept(ctx)
		if err != nil {
			return fmt.Errorf("accepting quic request error: %w", err)
		}

		config.LogDebug("QUIC connection established", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())

		select {
		case <-ctx.Done():
			goto done
		default:
			errGroup.Go(func() error {
				err := au.auth(ctx, conn, true)
				if err != nil {
					config.LogErr("[QUIC] authentication failed: %w", err)
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

func (au *authentication) auth(ctx context.Context, conn repository.ConnectionRepository, isQUIC bool) error {
	res := au.packetHandlerRepository.GenerateLoginResponse()
	nodeID := make([]byte, entity.IDlen)

	defer func() {
		config.LogDebug("Generated Login Response", "transaction_id", res.BaseHeader.TransactionID, "login_response", res)

		b, err := res.ConvertBytes()
		if err != nil {
			config.LogErr("Failed to convert to bytes", "error", err)
		}

		if _, err := conn.Write(b); err != nil {
			config.LogErr("Failed to write", "error", err)
		} else {
			config.LogDebug("Sent Login Response", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())
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
		au.packetHandlerRepository.ChangeStatus(res, entity.StatusClassInternalServerError)

		return fmt.Errorf("read error: %w", err)
	}

	req, shouldRead, err := au.packetHandlerRepository.Unmarshal(ctx, buf, length)
	if err != nil {
		au.packetHandlerRepository.ChangeStatus(res, entity.StatusClassInternalServerError)

		return fmt.Errorf("login request unmarshal error: %w", err)
	}

	if shouldRead {
		if length, err = conn.Read(buf); err != nil {
			au.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError, nodeID)

			return fmt.Errorf("failed to read until certificate error: %w", err)
		}

		au.packetHandlerRepository.ReCreateLoginRequest(ctx, req, buf, length)
	}

	config.LogDebug("Unmarshaled login request", "transaction_id", req.BaseHeader.TransactionID, "login_request", req)

	serverCert := au.certficateStoreRepository.Get(entity.ServerCert)
	rootCert := au.certficateStoreRepository.Get(entity.RootCert)

	device, status, err := au.sqlHandlerRepository.GetDevice(ctx, *req, serverCert, rootCert)
	if err != nil {
		au.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, status, nodeID)

		return fmt.Errorf("get device error: %w", err)
	}

	virtualIPAddress, err := au.sqlHandlerRepository.GetVirtualIPAddress(device)
	if err != nil {
		au.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError, nodeID)

		return fmt.Errorf("get virtual ip address error: %w", err)
	}

	nodeID, err = au.idGenerator.GenerateNodeID(device.FQDN, req.BaseHeader.ID)
	if err != nil {
		au.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError, nodeID)

		return fmt.Errorf("generate node id error: %w", err)
	}

	err = au.sqlHandlerRepository.SetNodeInformation(ctx, device, virtualIPAddress, req, nodeID)
	if err != nil {
		au.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError, nodeID)

		return fmt.Errorf("set node information error: %w", err)
	}

	accessToken, err := au.tokenGenerator.GenerateAccessToken(nodeID)
	if err != nil {
		au.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassInternalServerError, nodeID)

		return fmt.Errorf("generate access token error: %w", err)
	}

	au.packetHandlerRepository.SerializeSecretCommonValue(res)
	au.packetHandlerRepository.SerializePSInformation(res, isQUIC)
	au.packetHandlerRepository.SerializeCMSInformation(res, isQUIC)
	au.packetHandlerRepository.SerializeAccessToken(res, accessToken)
	au.packetHandlerRepository.SerializeBaseHeader(res, req.BaseHeader, entity.StatusClassSuccess, nodeID)

	return nil
}
