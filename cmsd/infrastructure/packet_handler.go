//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/cmsd/$GOPACKAGE
package infrastructure

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"

	"github.com/Pluslab/cyphonic/cmsd/entity"
	"github.com/Pluslab/cyphonic/cmsd/usecase/repository"
)

type packetHandler struct{}

func NewPacketHandler() repository.PacketHandler {
	return &packetHandler{}
}

func (ph *packetHandler) Unmarshal(ctx context.Context, buf []byte, length int) (*entity.ConnectionRequest, error) {
	if length < entity.BaseConnectionRequestLength {
		return nil, fmt.Errorf("invalid connection request packet length")
	}

	if buf[5] == uint8(entity.TypeClassConnectionRequest) {

		req := entity.ConnectionRequest{
			BaseHeader: entity.BaseHeader{
				TransactionID:  binary.BigEndian.Uint32(buf[0:4]),
				Version:        buf[4],
				Type:           buf[5],
				Status:         buf[6],
				Count:          buf[7],
				SequenceNumber: binary.BigEndian.Uint32(buf[8:12]),
				MessageLength:  binary.BigEndian.Uint16(buf[12:14]),
				Token:          buf[14],
				NextOpt:        buf[15],
				ID:             buf[16:32],
			},
			AccessTokenLength: binary.BigEndian.Uint16(buf[32:34]),
			Padding:           0,
			AccessToken:       "",
		}

		req.AccessToken = string(buf[entity.BaseConnectionRequestLength : entity.BaseConnectionRequestLength+req.AccessTokenLength])

		return &req, nil
	}

	return nil, fmt.Errorf("invalid packet type of connection request")
}

// func (ph *packetHandler) ReCreateConnectionRequest(ctx context.Context, req *entity.ConnectionRequest, buf []byte, length int) *entity.ConnectionRequest {
// 	req.Certificate = append(req.Certificate, buf[:length]...)

// 	return req
// }

func (ph *packetHandler) GenerateConnectionResponse() *entity.ConnectionResponse {
	return &entity.ConnectionResponse{
		BaseHeader: entity.BaseHeader{
			TransactionID:  0,
			Version:        0,
			Type:           uint8(entity.TypeClassConnectionResponse),
			Status:         0,
			Count:          0,
			SequenceNumber: 0,
			MessageLength:  0,
			Token:          0,
			NextOpt:        0,
			ID:             make([]byte, entity.IDlen),
		},
		NSIPv4Address: netip.AddrFrom4([4]byte(net.IPv4zero.To4())),
		NSIPv6Address: netip.AddrFrom16([16]byte(net.IPv6zero.To16())),
		NSPort:        0,
		Padding:       0,
		ExpireDate: entity.ExpireDate{
			Year:  0,
			Month: 0,
			Day:   0,
		},
		CipherType:      0,
		CommonKeyLength: 0,
		CommonKey:       []byte{},
	}
}

func (ph *packetHandler) SerializeBaseHeader(res *entity.ConnectionResponse, prevBaseHeader entity.BaseHeader, st entity.StatusClass) {
	res.SerializeBaseHeader(prevBaseHeader, entity.TypeClassConnectionResponse, st)
}

func (ph *packetHandler) SerializeNSInformation(res *entity.ConnectionResponse, nsInfo *entity.NotificationServiceInfomation) {
	res.SerializeNSInformation(nsInfo)
}

func (ph *packetHandler) SerializeCommonKey(res *entity.ConnectionResponse, commonKey *entity.CommonKey) {
	res.SerializeCommonKey(commonKey)
}

func (ph *packetHandler) ChangeStatus(res *entity.ConnectionResponse, st entity.StatusClass) {
	res.ChangeStatus(st)
}
