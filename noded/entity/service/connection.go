package service

import (
	"fmt"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ ConnectionService = (*connectionService)(nil)

type ConnectionService interface {
	Connection(loginResponse entity.LoginResponse) (*entity.ConnectionResponse, error)
}

type connectionService struct {
	repository repository.ConnectionRepository
}

func NewConnectionService(repository repository.ConnectionRepository) *connectionService {
	return &connectionService{
		repository: repository,
	}
}

func (s *connectionService) Connection(loginResponse entity.LoginResponse) (*entity.ConnectionResponse, error) {
	connectionResponse, err := s.repository.Connection(loginResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to connection: %w", err)
	}

	return connectionResponse, nil
}
