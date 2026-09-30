package authenticationservice

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

const (
	readBufferSize = 4096
	failedToLogin  = "failed to login"
)

var ErrFailedLogin = errors.New("failed to login")

var _ repository.AuthRepository = (*AuthenticationServiceRepository)(nil)

type AuthenticationServiceRepository struct {
	authenticationServiceAddress string
	tlsConfig                    *tls.Config
}

func NewAuthenticationServiceRepository(authenticationServiceAddress string, tlsConfig *tls.Config) *AuthenticationServiceRepository {
	return &AuthenticationServiceRepository{
		authenticationServiceAddress: authenticationServiceAddress,
		tlsConfig:                    tlsConfig,
	}
}

func (r *AuthenticationServiceRepository) Login(username, password string) (string, *entity.LoginResponse, error) {
	loginRequest, err := entity.GenerateLoginRequest(username, password)
	if err != nil {
		return failedToLogin, nil, fmt.Errorf("failed to generate login request: %w", err)
	}

	buffer, err := loginRequest.Marshal()
	if err != nil {
		return failedToLogin, nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	tlsConn, err := tls.Dial("tcp", r.authenticationServiceAddress, r.tlsConfig)
	if err != nil {
		return failedToLogin, nil, fmt.Errorf("failed to dial: %w", err)
	}

	defer tlsConn.Close()

	_, err = tlsConn.Write(buffer)
	if err != nil {
		return failedToLogin, nil, fmt.Errorf("failed to write login request: %w", err)
	}

	config.LogDebug("Sent login request to Authentication Service", "dst", r.authenticationServiceAddress, "username", username)

	readBuffer := make([]byte, readBufferSize)

	bufferLength, err := tlsConn.Read(readBuffer)
	if err != nil {
		return failedToLogin, nil, fmt.Errorf("failed to read login response: %w", err)
	}

	loginResponse, err := entity.UnmarshalLoginResponse(readBuffer[:bufferLength])
	if err != nil {
		return failedToLogin, nil, fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	config.LogDebug("Received login response from Authentication Service", "src", r.authenticationServiceAddress, "status", loginResponse.BaseHeader.Status)

	if loginResponse.BaseHeader.Status != entity.StatusClassSuccess {
		return failedToLogin, nil, ErrFailedLogin
	}

	return "login successful!", loginResponse, nil
}
