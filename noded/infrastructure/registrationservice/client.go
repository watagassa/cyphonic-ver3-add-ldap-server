package registrationservice

import (
	"fmt"
	"net/netip"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

// readBufferSize is the size of the buffer used to read from the udp connection.
const readBufferSize = 4096

var _ repository.RegistrationRepository = (*RegistrationServiceRepository)(nil)

type RegistrationServiceRepository struct {
	socket entity.Socket
}

func NewRegistrationServiceRepository(socket entity.Socket) *RegistrationServiceRepository {
	return &RegistrationServiceRepository{
		socket: socket,
	}
}

func (r *RegistrationServiceRepository) Registration(
	loginResponse entity.LoginResponse,
	nodeIPv4Address, nodeIPv6Address netip.Addr,
	nsAddress *entity.NSAddress, interfaceName string,
) (*entity.RegistrationResponse, error) {
	registrationRequest := entity.GenerateRegistrationRequest(loginResponse.BaseHeader, nodeIPv4Address, nodeIPv6Address, interfaceName)

	buffer, err := registrationRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	localIPVersion := entity.DetermineLocalIPVersion(nodeIPv4Address, nodeIPv6Address)
	nsAddr, err := nsAddress.UDPAddr(localIPVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to determine NS address: %w", err)
	}

	// TODO: 送信前にcommonKeyで暗号化しないといけない
	_, err = r.socket.WriteToUDP(buffer, &nsAddr)

	if err != nil {
		return nil, fmt.Errorf("failed to write registration request: %w", err)
	}

	config.LogDebug("Sent registration request to Registration Service",
		"src", r.socket.LocalAddr(),
		"dst", nsAddr.String())

	readBuffer := make([]byte, readBufferSize)
	bufferLength, _, err := r.socket.ReadFromUDP(readBuffer)

	if err != nil {
		return nil, fmt.Errorf("failed to read registration response: %w", err)
	}

	registrationResponse, err := entity.UnmarshalRegistrationResponse(readBuffer[:bufferLength])
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal registration response: %w", err)
	}

	config.LogDebug("Received registration response from Registration Service", "src", nsAddr.String(), "dst", r.socket.LocalAddr())

	return registrationResponse, nil
}
