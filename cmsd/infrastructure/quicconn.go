package infrastructure

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/cmsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/cmsd/usecase/repository"
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

func (q *quicListener) Accept(ctx context.Context) (repository.ConnectionRepository, error) {
	conn, err := q.QUICListener.Accept(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't accept quic listen: %w", err)
	}

	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		switch err.Error() {
		case "Application error 0x0 (remote)":
		default:
			return nil, err
		}
	}

	return &QUICStream{
		conn: conn,
		st:   stream,
	}, nil
}

func (q *quicListener) Close() error {
	return fmt.Errorf("can't close quic listen: %w", q.QUICListener.Close())
}

type QUICStream struct {
	conn quic.Connection
	st   quic.Stream
}

func (q *QUICStream) LocalAddr() string {
	return q.conn.LocalAddr().String()
}

func (q *QUICStream) RemoteAddr() string {
	return q.conn.RemoteAddr().String()
}

func (q *QUICStream) Read(buf []byte) (int, error) {
	length, err := q.st.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("stream read error in quic stream: %w", err)
	}

	return length, nil
}

func (q *QUICStream) Write(buf []byte) (int, error) {
	if length, err := q.st.Write(buf); err != nil {
		return 0, fmt.Errorf("stream write error in quic stream: %w", err)
	} else {
		return length, nil
	}
}

func (q *QUICStream) Close() error {
	if err := q.st.Close(); err != nil {
		return fmt.Errorf("failed to close quic connection: %w", err)
	}

	return nil
}
