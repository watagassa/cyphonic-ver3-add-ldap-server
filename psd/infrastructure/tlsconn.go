package infrastructure

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"

	entity "github.com/Pluslab/cyphonic/psd/entity"
	"github.com/Pluslab/cyphonic/psd/infrastructure/config"
	"github.com/Pluslab/cyphonic/psd/usecase/repository"
)

type tlsListener struct {
	TLSListener net.Listener
}

func NewTLSListener(cfg *config.Config, certStore repository.CertificateStore) (repository.TLSListener, error) {
	tlsConf, err := getTLSConfig(certStore, false)
	if err != nil {
		return nil, fmt.Errorf("can't get certificate: %w", err)
	}

	tlsListen, err := tls.Listen("tcp", fmt.Sprintf(":%s", cfg.TLSPort), tlsConf)
	if err != nil {
		return nil, fmt.Errorf("can't listen TLS: %w", err)
	}

	return &tlsListener{
		TLSListener: tlsListen,
	}, nil
}

func getTLSConfig(certStore repository.CertificateStore, isQUIC bool) (*tls.Config, error) {
	serverCert := certStore.Get(entity.ServerCert)
	serverCertPrivKey := certStore.Get(entity.ServerCertPrivKey)

	cert, err := tls.X509KeyPair(serverCert, serverCertPrivKey)
	if err != nil {
		return nil, fmt.Errorf("can't load x509 key pair: %w", err)
	}

	tlsConf := &tls.Config{
		Rand:                        nil,
		Time:                        nil,
		Certificates:                []tls.Certificate{cert},
		NameToCertificate:           map[string]*tls.Certificate{},
		GetCertificate:              nil,
		GetClientCertificate:        nil,
		GetConfigForClient:          nil,
		VerifyPeerCertificate:       nil,
		VerifyConnection:            nil,
		RootCAs:                     nil,
		NextProtos:                  []string{},
		ServerName:                  "",
		ClientAuth:                  tls.NoClientCert,
		ClientCAs:                   nil,
		InsecureSkipVerify:          false,
		CipherSuites:                []uint16{},
		PreferServerCipherSuites:    true,
		SessionTicketsDisabled:      false,
		SessionTicketKey:            [32]byte{},
		ClientSessionCache:          nil,
		MinVersion:                  tls.VersionTLS13,
		MaxVersion:                  0,
		CurvePreferences:            []tls.CurveID{},
		DynamicRecordSizingDisabled: false,
		Renegotiation:               0,
		KeyLogWriter:                nil,
	}

	if isQUIC {
		CAPool := x509.NewCertPool()

		caCert, err := os.ReadFile(certStore.GetRootCertificateName())
		if err != nil {
			return nil, fmt.Errorf("can't load ca certificate: %w", err)
		}

		if ok := CAPool.AppendCertsFromPEM(caCert); !ok {
			return nil, errors.New("certificate is not correct")
		}

		tlsConf.NextProtos = []string{"quic"}
		tlsConf.ClientCAs = CAPool
	}

	return tlsConf, nil
}

func (t *tlsListener) Accept() (repository.ConnectionRepository, error) {
	conn, err := t.TLSListener.Accept()
	if err != nil {
		return nil, fmt.Errorf("can't accept tls listen: %w", err)
	}

	bufioReadWriter := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter((conn)))

	return &TLSConnection{
		conn:       conn,
		readWriter: bufioReadWriter,
	}, nil
}

func (t *tlsListener) Close() error {
	return fmt.Errorf("can't close tls listen: %w", t.TLSListener.Close())
}

type TLSConnection struct {
	conn       net.Conn
	readWriter *bufio.ReadWriter
}

func (t *TLSConnection) LocalAddr() string {
	return t.conn.LocalAddr().String()
}

func (t *TLSConnection) RemoteAddr() string {
	return t.conn.RemoteAddr().String()
}

func (t *TLSConnection) Read(buf []byte) (int, error) {
	length, err := t.readWriter.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("bufio read error in tls connection: %w", err)
	}

	return length, nil
}

func (t *TLSConnection) Write(buf []byte) (int, error) {
	var n int
	var err error

	if n, err = t.readWriter.Write(buf); err != nil {
		return 0, fmt.Errorf("bufio write error: %w", err)
	} else if err = t.readWriter.Flush(); err != nil {
		return 0, fmt.Errorf("bufio flush error: %w", err)
	}

	return n, nil
}

func (t *TLSConnection) Close() error {
	if err := t.conn.Close(); err != nil {
		return fmt.Errorf("failed to close tls connection: %w", err)
	}

	return nil
}
