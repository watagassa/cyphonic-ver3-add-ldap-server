package finalizationservice

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

// readBufferSize is the size of the buffer used to read from the udp connection.
const readBufferSize = 4096

var _ repository.FinalizationRepository = (*FinalizationServiceRepository)(nil)

type FinalizationServiceRepository struct {
	tlsConfig *tls.Config
}

func NewFinalizationServiceRepository(tlsConfig *tls.Config) *FinalizationServiceRepository {
	return &FinalizationServiceRepository{
		tlsConfig: tlsConfig,
	}
}

func (r *FinalizationServiceRepository) Finalization(loginResponse entity.LoginResponse) (*entity.FinalizationResponse, error) {
	finalizationRequest := entity.GenerateFinalizationRequest(loginResponse.BaseHeader, loginResponse.AccessToken)

	buffer, err := finalizationRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal finalization request: %w", err)
	}

	// TODO: FQDN直打ち
	finalizationServiceAddress := net.JoinHostPort("fs.local.cyphonic.org", "4510")
	// finalizationServiceAddress := net.JoinHostPort(str(loginResponse.FSFQDN), strconv.Itoa(int(loginResponse.FSPort)))

	tlsConn, err := tls.Dial("tcp", finalizationServiceAddress, r.tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	defer tlsConn.Close()

	_, err = tlsConn.Write(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to write finalization request: %w", err)
	}

	config.LogDebug("Sent finalization request to Finalization Service", "dst", finalizationServiceAddress)

	readBuffer := make([]byte, readBufferSize)

	bufferLength, err := tlsConn.Read(readBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read finalization response: %w", err)
	}

	finalizationResponse, err := entity.UnmarshalFinalizationResponse(readBuffer[:bufferLength])
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ack from finalization service: %w", err)
	}

	config.LogDebug("Received ACK from Finalization Service", "src", finalizationServiceAddress)

	return finalizationResponse, nil
}
