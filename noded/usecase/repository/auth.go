//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/noded/$GOPACKAGE
package repository

import "github.com/Pluslab/cyphonic/noded/entity"

type AuthRepository interface {
	Login(username, password string) (string, *entity.LoginResponse, error)
}
