package repository

import "github.com/Pluslab/cyphonic/cmsd/entity"

type TokenVerifier interface {
	VerifyAccessToken(string, entity.ID) error
}
