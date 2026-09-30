package service

import (
	"fmt"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ AuthService = (*authService)(nil)

type AuthService interface {
	Login(username, password string) (string, *entity.LoginResponse, error)
}

type authService struct {
	repository repository.AuthRepository
}

func NewAuthService(repository repository.AuthRepository) *authService {
	return &authService{
		repository: repository,
	}
}

func (s *authService) Login(username, password string) (string, *entity.LoginResponse, error) {
	// TODO: Implement Login method
	message, loginResponse, err := s.repository.Login(username, password)
	if err != nil {
		return "failed to login", nil, fmt.Errorf("failed to login: %w", err)
	}

	return message, loginResponse, nil
}
