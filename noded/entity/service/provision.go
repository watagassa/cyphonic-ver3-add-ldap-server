package service

import (
	"fmt"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ ProvisionService = (*provisionService)(nil)

type ProvisionService interface {
	Provision(loginResponse entity.LoginResponse, desiredFQDN []byte) (*entity.ProvisionResponse, error)
}

type provisionService struct {
	repository repository.ProvisionRepository
}

func NewProvisionService(repository repository.ProvisionRepository) *provisionService {
	return &provisionService{
		repository: repository,
	}
}

func (s *provisionService) Provision(loginResponse entity.LoginResponse, desiredFQDN []byte) (*entity.ProvisionResponse, error) {
	provisionResponse, err := s.repository.Provision(loginResponse, desiredFQDN)
	if err != nil {
		return nil, fmt.Errorf("failed to provision: %w", err)
	}

	return provisionResponse, nil
}
