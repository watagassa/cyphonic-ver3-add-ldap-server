//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/asd/$GOPACKAGE
package repository

import (
	"context"

	"github.com/Pluslab/cyphonic/asd/entity"
)

type PacketHandler interface {
	Unmarshal(context.Context, []byte, int) (*entity.LoginRequest, bool, error)
	ReCreateLoginRequest(context.Context, *entity.LoginRequest, []byte, int) *entity.LoginRequest
	GenerateLoginResponse() *entity.LoginResponse
	SerializeBaseHeader(*entity.LoginResponse, entity.BaseHeader, entity.StatusClass, []byte)
	SerializeSecretCommonValue(*entity.LoginResponse)
	SerializePSInformation(*entity.LoginResponse, bool)
	SerializeCMSInformation(*entity.LoginResponse, bool)
	SerializeAccessToken(*entity.LoginResponse, string)
	ChangeStatus(*entity.LoginResponse, entity.StatusClass)
}
