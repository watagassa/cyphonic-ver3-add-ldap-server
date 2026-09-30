package service

import (
	"fmt"
	"net/netip"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ RegistrationService = (*registrationService)(nil)

type RegistrationService interface {
	Registration(loginResponse entity.LoginResponse,
		nodeIPv4Address, nodeIPv6Address netip.Addr,
		nsAddress *entity.NSAddress, interfaceName string) (*entity.RegistrationResponse, error)
}

type registrationService struct {
	repository repository.RegistrationRepository
}

func NewRegistrationService(repository repository.RegistrationRepository) *registrationService {
	return &registrationService{
		repository: repository,
	}
}

func (s *registrationService) Registration(
	loginResponse entity.LoginResponse,
	nodeIPv4Address, nodeIPv6Address netip.Addr,
	nsAddress *entity.NSAddress,
	interfaceName string,
) (*entity.RegistrationResponse, error) {
	registrationResponse, err := s.repository.Registration(loginResponse, nodeIPv4Address, nodeIPv6Address, nsAddress, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to registration: %w", err)
	}

	return registrationResponse, nil
}
