//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/cmsd/$GOPACKAGE
package repository

import (
	"github.com/Pluslab/cyphonic/cmsd/entity"
)

type SQLHandler interface {
	GetRandomNotificationService() (*entity.NotificationServiceInfomation, entity.StatusClass, error)
	SetNodeInformation(entity.ConnectionRequest, string, entity.ExpireDate, entity.TypeCipherClass) error
}
