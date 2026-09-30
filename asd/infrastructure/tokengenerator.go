//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/asd/$GOPACKAGE
package infrastructure

import (
	"fmt"
	"time"

	"github.com/Pluslab/cyphonic/asd/infrastructure/config"
	jwt "github.com/golang-jwt/jwt/v4"
)

type tokenGenerator struct {
	accessSecretKey string
}

func NewTokenGenerator(cfg *config.Config) *tokenGenerator {
	return &tokenGenerator{
		accessSecretKey: cfg.AccessSecretKey,
	}
}

func (tg *tokenGenerator) GenerateAccessToken(nodeID []byte) (string, error) {
	claims := jwt.MapClaims{
		"node_id": nodeID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(tg.accessSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}
