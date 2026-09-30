package infrastructure

import (
	"fmt"
	"strings"
	"time"

	"github.com/Pluslab/cyphonic/cmsd/entity"
	"github.com/Pluslab/cyphonic/cmsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/cmsd/usecase/repository"
	"github.com/google/uuid"
)

type commonKeyGenerator struct {
}

func NewCommonKeyGenerator(cfg *config.Config) repository.CommonKeyGenerator {
	return &commonKeyGenerator{}
}

func (cg *commonKeyGenerator) GenerateCommonKey() (*entity.CommonKey, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("random number creation error: %w", err)
	}

	commonKey := strings.Replace(u.String(), "-", "", -1)
	expireDate := time.Now().AddDate(0, 1, 0)

	return &entity.CommonKey{
		CommonKey: commonKey,
		ExpireDate: entity.ExpireDate{
			Year:  uint16(expireDate.Year()),
			Month: uint8(expireDate.Month()),
			Day:   uint8(expireDate.Day()),
		},
		// FIXME: 暗号タイプが固定になっている
		CipherType: entity.AES256CBC,
	}, nil
}
