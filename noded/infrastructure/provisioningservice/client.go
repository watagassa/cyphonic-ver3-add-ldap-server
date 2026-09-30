package provisioningservice

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

var _ repository.ProvisionRepository = (*ProvisionServiceRepository)(nil)

type ProvisionServiceRepository struct {
	tlsConfig *tls.Config
}

func NewProvisionServiceRepository(tlsConfig *tls.Config) *ProvisionServiceRepository {
	return &ProvisionServiceRepository{
		tlsConfig: tlsConfig,
	}
}

func (r *ProvisionServiceRepository) Provision(loginResponse entity.LoginResponse, desiredFQDN []byte) (*entity.ProvisionResponse, error) {
	provisionRequest := entity.GenerateProvisionRequest(loginResponse.BaseHeader, desiredFQDN, loginResponse.AccessToken)

	buffer, err := provisionRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal provision request: %w", err)
	}

	provisioningServiceAddress := net.JoinHostPort(string(loginResponse.PSFQDN), strconv.Itoa(int(loginResponse.PSPort)))

	tlsConn, err := tls.Dial("tcp", provisioningServiceAddress, r.tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	defer tlsConn.Close()

	_, err = tlsConn.Write(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to write provision request: %w", err)
	}

	config.LogDebug("Sent provision request to Provisioning Service", "dst", provisioningServiceAddress)

	readBuffer := make([]byte, readBufferSize)

	bufferLength, err := tlsConn.Read(readBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read provision response: %w", err)
	}

	provisionResponse, err := entity.UnmarshalProvisionResponse(readBuffer[:bufferLength])
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal provision response: %w", err)
	}

	config.LogDebug("Received provision response from Provisioning Service",
		"src", provisioningServiceAddress,
		"virtual_ipv4", provisionResponse.VirtualIPv4Address,
		"virtual_ipv6", provisionResponse.VirtualIPv6Address,
		"fqdn", provisionResponse.FQDN,
	)

	return provisionResponse, nil
}
