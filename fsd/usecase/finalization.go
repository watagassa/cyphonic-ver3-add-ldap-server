//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/fsd/$GOPACKAGE
package usecase

import (
	"bufio"
	"context"
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/fsd/entity"
	"github.com/Pluslab/cyphonic/fsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/fsd/usecase/repository"
	quic "github.com/quic-go/quic-go"
	"golang.org/x/sync/errgroup"
)

const BufferSize = 4096

type Finalization interface {
	Finalization(ctx context.Context) error
}

type finalization struct {
	packetHandlerRepository   repository.PacketHandler
	sqlHandlerRepository      repository.SQLHandler
	certficateStoreRepository repository.CertificateStore
	tlsListenerRepository     repository.TLSListener
	quicListenerRepository    repository.QUICListener
}

func NewFinalization(ph repository.PacketHandler, sh repository.SQLHandler, cs repository.CertificateStore, tl repository.TLSListener, ql repository.QUICListener) Finalization {
	return &finalization{
		packetHandlerRepository:   ph,
		sqlHandlerRepository:      sh,
		certficateStoreRepository: cs,
		tlsListenerRepository:     tl,
		quicListenerRepository:    ql,
	}
}

func (fi *finalization) Finalization(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return fi.listenAndServeOverTLS(ctx, errGroup)
	})
	errGroup.Go(func() error {
		return fi.listenAndServeOverQUIC(ctx, errGroup)
	})

	config.LogInfo("Finalizartion service is running...")

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("error occurred: %s", err)
	}

	return nil
}

func (fi *finalization) listenAndServeOverTLS(ctx context.Context, errGroup *errgroup.Group) error {
	config.LogInfo("Activating TLS server")

	for {
		conn, err := fi.tlsListenerRepository.Accept()
		if err != nil {
			return fmt.Errorf("accepting tls request error: %s", err)
		}

		config.LogInfo("TLS connection established", "src", conn.LocalAddr().String(), "dst", conn.RemoteAddr().String())

		select {
		case <-ctx.Done():
			goto done
		default:
			errGroup.Go(func() error {
				return fi.handleTLSConn(ctx, conn)
			})
		}
	}

done:
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("errGroup detected error in TLS server: %w", err)
	}

	return nil
}

func (fi *finalization) handleTLSConn(ctx context.Context, conn net.Conn) error {
	defer func() {
		if err := conn.Close(); err != nil {
			err := fmt.Errorf("failed to close tls connection: %w", err)
			config.LogErr("Error occurred when closing TLS connection", "error", err)
		}

		config.LogInfo("TLS connection closed", "src", conn.LocalAddr().String(), "dst", conn.RemoteAddr().String())
	}()

	buf := make([]byte, BufferSize)
	bufioReadWriter := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter((conn)))

	length, err := bufioReadWriter.Read(buf)
	if err != nil {
		err = fmt.Errorf("bufio read error: %w", err)
		config.LogErr("Failed to read from bufio in TLS connection: %w", "error", err)

		return err
	}

	req, shouldRead, err := fi.packetHandlerRepository.Unmarshal(ctx, buf, length)
	if err != nil {
		err = fmt.Errorf("finalization request unmarshal error: %w", err)
		config.LogErr("Failed to unmarshal finalization request in TLS connection: %w", "error", err)

		return err
	}

	if shouldRead {
		if length, err = bufioReadWriter.Read(buf); err != nil {
			err = fmt.Errorf("bufio read until certificate error: %w", err)
			config.LogErr("Failed to read certificate from bufio", "error", err)

			return err
		}
	}

	config.LogDebug("Unmarshaled finalization request", "transaction_id", req.BaseHeader.TransactionID, "finalization_request", req)

	res, err := fi.handleFinalizationRequest(ctx, req)
	if err != nil {
		err = fmt.Errorf("handling finalization request error: %w", err)
		config.LogErr("Failed to handle finalization request in TLS connection", "error", err)

		return err
	}

	b, err := res.ConvertBytes()
	if err != nil {
		err = fmt.Errorf("convert bytes error: %w", err)
		config.LogErr("Failed to convert to bytes", "error", err)

		return err
	}

	if _, ok := bufioReadWriter.Write(b); ok != nil {
		err = fmt.Errorf("bufio write error: %w", ok)
		config.LogErr("Failed to write to bufio in TLS conection", "error", err)

		return err
	} else if err = bufioReadWriter.Flush(); err != nil {
		err = fmt.Errorf("bufio flush error: %w", ok)
		config.LogErr("Failed to flush bufio in TLS connection", "error", err)

		return err
	}

	config.LogDebug("Sent Finalization Response", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())

	return nil
}

func (fi *finalization) listenAndServeOverQUIC(ctx context.Context, errGroup *errgroup.Group) error {
	config.LogInfo("Activating QUIC server")

	for {
		conn, err := fi.quicListenerRepository.Accept(ctx)
		if err != nil {
			return fmt.Errorf("accepting quic request error: %w", err)
		}

		config.LogInfo("QUIC connection established", "src", conn.LocalAddr().String(), "dst", conn.RemoteAddr().String())

		select {
		case <-ctx.Done():
			goto done
		default:
			errGroup.Go(func() error {
				return fi.handleQUICConn(ctx, conn)
			})
		}
	}

done:
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("errGroup detected error in QUIC server: %w", err)
	}

	return nil
}

func (fi *finalization) handleQUICConn(ctx context.Context, conn quic.Connection) error {
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		switch err.Error() {
		case "Application error 0x0 (remote)":
		default:
			return err
		}
	}

	defer func() {
		if err := stream.Close(); err != nil {
			err := fmt.Errorf("failed to close quic connection: %w", err)
			config.LogErr("Error occurred when closing QUIC connection", "error", err)
		}

		config.LogInfo("QUIC connection closed", "src", conn.LocalAddr().String(), "dst", conn.RemoteAddr().String())
	}()

	buf := make([]byte, BufferSize)

	length, err := stream.Read(buf)
	if err != nil {
		err = fmt.Errorf("stream read error: %w", err)
		config.LogErr("Failed to read from stream in QUIC connection: %w", "error", err)

		return err
	}

	req, shouldRead, err := fi.packetHandlerRepository.Unmarshal(ctx, buf, length)
	if err != nil {
		err = fmt.Errorf("finalization request unmarshal error: %w", err)
		config.LogErr("Failed to unmarshal finalization request in QUIC connection: %w", "error", err)

		return err
	}

	if shouldRead {
		if length, err = stream.Read(buf); err != nil {
			err = fmt.Errorf("stream read until certificate error: %w", err)
			config.LogErr("Failed to read certificate from stream", "error", err)

			return err
		}
	}

	config.LogDebug("Unmarshaled finalization request", "transaction_id", req.BaseHeader.TransactionID, "finalization_request", req)

	res, err := fi.handleFinalizationRequest(ctx, req)
	if err != nil {
		err = fmt.Errorf("handling finalization request error: %w", err)
		config.LogErr("Failed to handle login request in TLS connection", "error", err)

		return err
	}

	b, err := res.ConvertBytes()
	if err != nil {
		err = fmt.Errorf("convert bytes error: %w", err)
		config.LogErr("Failed to convert to bytes", "error", err)

		return err
	}

	if _, ok := stream.Write(b); ok != nil {
		err = fmt.Errorf("stream write error: %w", ok)
		config.LogErr("Failed to write to stream", "error", err)

		return err
	}

	config.LogDebug("Sent Finalization Response", "src", conn.LocalAddr(), "dst", conn.RemoteAddr())

	return nil
}

func (fi *finalization) handleFinalizationRequest(ctx context.Context, req *entity.FinalizationRequest) (*entity.FinalizationResponse, error) {

	if err := fi.sqlHandlerRepository.DeleteNodeInformation(ctx, *req); err != nil {
		err = fmt.Errorf("delete node information error: %w", err)
		config.LogErr("Failed to delete node information", "error", err)

		return nil, err
	}

	res, err := fi.packetHandlerRepository.GenerateFinalizationResponse(req.BaseHeader)
	if err != nil {
		err = fmt.Errorf("create finalization response error: %w", err)
		config.LogErr("Failed to generate finalization response", "error", err)

		return nil, err
	}

	config.LogDebug("Generated Finalization Response", "transaction_id", res.BaseHeader.TransactionID, "finalization_response", res)

	return res, nil
}
