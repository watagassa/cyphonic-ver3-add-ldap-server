//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/noded/$GOPACKAGE
package repository

import (
	"net/netip"

	"github.com/Pluslab/cyphonic/noded/entity"
)

type RegistrationRepository interface {
	Registration(loginResponse entity.LoginResponse,
		nodeIPv4Address, nodeIPv6Address netip.Addr,
		nsAddress *entity.NSAddress, interfaceName string) (*entity.RegistrationResponse, error)
}
