package connectionservice

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

// readBufferSize is the size of the buffer used to read from the udp connection.
const readBufferSize = 4096

var _ repository.ConnectionRepository = (*ConnectionServiceRepository)(nil)

type ConnectionServiceRepository struct {
	tlsConfig *tls.Config
}

func NewConnectionServiceRepository(tlsConfig *tls.Config) *ConnectionServiceRepository {
	return &ConnectionServiceRepository{
		tlsConfig: tlsConfig,
	}
}

func (r *ConnectionServiceRepository) Connection(loginResponse entity.LoginResponse) (*entity.ConnectionResponse, error) {
	connectionRequest := entity.GenerateConnectionRequest(loginResponse.BaseHeader, loginResponse.AccessToken)

	buffer, err := connectionRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal connection request: %w", err)
	}

	connectionServiceAddress := net.JoinHostPort(string(loginResponse.CMSFQDN), strconv.Itoa(int(loginResponse.CMSPort)))

	tlsConn, err := tls.Dial("tcp", connectionServiceAddress, r.tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	defer tlsConn.Close()

	_, err = tlsConn.Write(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to write connection request: %w", err)
	}

	config.LogDebug("Sent connection request to Connection Service", "dst", connectionServiceAddress)

	readBuffer := make([]byte, readBufferSize)

	bufferLength, err := tlsConn.Read(readBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read connection response: %w", err)
	}

	connectionResponse, err := entity.UnmarshalConnectionResponse(readBuffer[:bufferLength])
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal connection response: %w", err)
	}

	config.LogDebug("Received connection response from Connection Service",
		"dst", connectionServiceAddress,
		"ns_ipv4", connectionResponse.NSIPv4Address,
		"ns_ipv6", connectionResponse.NSIPv6Address,
		"ns_port", connectionResponse.NSPort,
	)

	return connectionResponse, nil
}
