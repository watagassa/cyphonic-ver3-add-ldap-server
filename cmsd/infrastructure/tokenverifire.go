//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/cmsd/$GOPACKAGE
package infrastructure

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Pluslab/cyphonic/cmsd/entity"
	"github.com/Pluslab/cyphonic/cmsd/infrastructure/config"
	jwt "github.com/golang-jwt/jwt/v4"
)

type tokenVerifier struct {
	accessSecretKey string
}

func NewTokenVerifire(cfg *config.Config) *tokenVerifier {
	return &tokenVerifier{
		accessSecretKey: cfg.AccessSecretKey,
	}
}

func (tv *tokenVerifier) VerifyAccessToken(tokenStr string, nodeID entity.ID) error {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(tv.accessSecretKey), nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		encodedNodeID := string(claims["node_id"].(string))
		exp := int64(claims["exp"].(float64))

		if time.Now().Unix() > exp {
			return errors.New("expired token")
		}

		decodedNodeID, err := base64.StdEncoding.DecodeString(encodedNodeID)
		if err != nil {
			return fmt.Errorf("failed to decode node id: %w", err)
		}

		if bytes.Equal(decodedNodeID, nodeID) {
			return nil
		} else {
			return errors.New("forbidden request")
		}
	} else {
		return fmt.Errorf("invalid token: %w", err)
	}
}
