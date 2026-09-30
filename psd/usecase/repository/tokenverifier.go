package repository

import "github.com/Pluslab/cyphonic/psd/entity"

type TokenVerifier interface {
	VerifyAccessToken(string, entity.ID) error
}
