package service

import (
	"fmt"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ FinalizationService = (*finalizationService)(nil)

type FinalizationService interface {
	Finalization(loginResponse entity.LoginResponse) (*entity.FinalizationResponse, error)
}

type finalizationService struct {
	repository repository.FinalizationRepository
}

func NewFinalizationService(repository repository.FinalizationRepository) *finalizationService {
	return &finalizationService{
		repository: repository,
	}
}

func (s *finalizationService) Finalization(loginResponse entity.LoginResponse) (*entity.FinalizationResponse, error) {
	finalizationResponse, err := s.repository.Finalization(loginResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to finalization: %w", err)
	}

	return finalizationResponse, nil
}
