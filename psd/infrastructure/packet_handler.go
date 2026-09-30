//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/psd/$GOPACKAGE
package infrastructure

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"

	"github.com/Pluslab/cyphonic/psd/entity"
	"github.com/Pluslab/cyphonic/psd/usecase/repository"
)

type packetHandler struct {
	vipVersion entity.VIPVersionClass
}

func NewPacketHandler(vipVersion int) (repository.PacketHandler, error) {
	if !(vipVersion == 4 || vipVersion == 6) {
		return nil, errors.New("illegal virtual ip version")
	}

	return &packetHandler{
		vipVersion: entity.VIPVersionClass(vipVersion),
	}, nil
}

func (ph *packetHandler) Unmarshal(ctx context.Context, buf []byte, length int) (*entity.ProvisionRequest, error) {
	if length < entity.BaseProvisionRequestLength {
		return nil, fmt.Errorf("invalid provision request packet length")
	}

	if buf[5] == uint8(entity.TypeClassProvisionRequest) {
		req := entity.ProvisionRequest{
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
				ID:             make(entity.ID, entity.IDlen),
			},
			DesiredFQDNLength: binary.BigEndian.Uint16(buf[32:34]),
			AccessTokenLength: binary.BigEndian.Uint16(buf[34:36]),
			DesiredFQDN:       "",
			AccessToken:       "",
		}

		copy(req.BaseHeader.ID, buf[16:32])
		req.DesiredFQDN = string(buf[entity.BaseProvisionRequestLength : entity.BaseProvisionRequestLength+req.DesiredFQDNLength])
		req.AccessToken = string(buf[entity.BaseProvisionRequestLength+req.DesiredFQDNLength : entity.BaseProvisionRequestLength+req.DesiredFQDNLength+req.AccessTokenLength])

		return &req, nil
	}

	return nil, fmt.Errorf("invalid packet type of provision request")
}

func (ph *packetHandler) GenerateProvisionResponse() *entity.ProvisionResponse {
	var ipv4zeroSlice [4]byte
	var ipv6zeroSlice [16]byte
	copy(ipv4zeroSlice[:], net.IPv4zero.To4())
	copy(ipv6zeroSlice[:], net.IPv6zero.To16())

	ipv4 := netip.AddrFrom4(ipv4zeroSlice)
	ipv6 := netip.AddrFrom16(ipv6zeroSlice)

	return &entity.ProvisionResponse{
		BaseHeader: entity.BaseHeader{
			TransactionID:  0,
			Version:        0,
			Type:           uint8(entity.TypeClassProvisionResponse),
			Status:         0,
			Count:          0,
			SequenceNumber: 0,
			MessageLength:  entity.BaseProvisionResponseLength,
			Token:          0,
			NextOpt:        0,
			ID:             make([]byte, entity.IDlen),
		},
		VirtualIPv4Address: ipv4,
		VirtualIPv6Address: ipv6,
		VirtualIPv4Prefix:  0,
		VirtualIPv6Prefix:  0,
		VIPVersion:         0,
		L2Flag:             0,
		FQDNLength:         0,
		FQDN:               "",
	}
}

func (ph *packetHandler) SerializeBaseHeader(res *entity.ProvisionResponse, prevBaseHeader entity.BaseHeader, st entity.StatusClass) {
	res.SerializeBaseHeader(prevBaseHeader, entity.TypeClassProvisionResponse, st)
}

func (ph *packetHandler) SerializeVirtualIPAddress(res *entity.ProvisionResponse, vip4, vip6 netip.Prefix) {
	res.SerializeVirtualIP(vip4, vip6, ph.vipVersion)
}

func (ph *packetHandler) SerializeL2Flag(res *entity.ProvisionResponse) {
	res.SerializeL2Flag()
}

func (ph *packetHandler) SerializeFQDN(res *entity.ProvisionResponse, fqdn string) {
	res.SerializeFQDN(fqdn)
}

func (ph *packetHandler) ChangeStatus(res *entity.ProvisionResponse, st entity.StatusClass) {
	res.ChangeStatus(st)
}
