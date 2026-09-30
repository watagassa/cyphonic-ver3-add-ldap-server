package infrastructure

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/fsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/fsd/usecase/repository"
	quic "github.com/quic-go/quic-go"
)

const connectionIDLength int = 16

type quicListener struct {
	QUICListener *quic.Listener
}

func NewQUICListener(cfg *config.Config, certStore repository.CertificateStore) (repository.QUICListener, error) {
	tlsConf, err := getTLSConfig(certStore, true)
	if err != nil {
		return nil, fmt.Errorf("can't get certificate: %w", err)
	}

	quicConf := getQUICConfig()

	quicListen, err := newQUICListen(cfg.QUICPort, tlsConf, quicConf)
	if err != nil {
		return nil, fmt.Errorf("can't listen QUIC: %w", err)
	}

	return &quicListener{
		QUICListener: quicListen,
	}, nil
}

func getQUICConfig() *quic.Config {
	quicConf := &quic.Config{
		Versions:        []quic.Version{quic.Version1},
		KeepAlivePeriod: 0,
		Tracer:          nil,
	}

	return quicConf
}

func newQUICListen(port string, tlsConf *tls.Config, quicConf *quic.Config) (*quic.Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve of udp address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to open an udp socket: %w", err)
	}

	qt := &quic.Transport{
		Conn:               udpConn,
		ConnectionIDLength: connectionIDLength,
	}

	return qt.Listen(tlsConf, quicConf)
}

func (q *quicListener) Accept(ctx context.Context) (quic.Connection, error) {
	conn, err := q.QUICListener.Accept(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't accept quic listen: %w", err)
	}

	return conn, nil
}

func (q *quicListener) Close() error {
	return fmt.Errorf("can't close quic listen: %w", q.QUICListener.Close())
}
