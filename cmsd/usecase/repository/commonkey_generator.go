package repository

import "github.com/Pluslab/cyphonic/cmsd/entity"

type CommonKeyGenerator interface {
	GenerateCommonKey() (*entity.CommonKey, error)
}
