//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/noded/$GOPACKAGE
package repository

import "github.com/Pluslab/cyphonic/noded/entity"

type FinalizationRepository interface {
	Finalization(loginResponse entity.LoginResponse) (*entity.FinalizationResponse, error)
}
