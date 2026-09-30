//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/asd/$GOPACKAGE
package infrastructure

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/Pluslab/cyphonic/asd/entity"
	"github.com/Pluslab/cyphonic/asd/infrastructure/config"
	"github.com/Pluslab/cyphonic/asd/usecase/repository"
)

type packetHandler struct {
	psFQDN      string
	psTLSPort   int
	psQUICPort  int
	cmsFQDN     string
	cmsTLSPort  int
	cmsQUICPort int
}

func NewPacketHandler(cfg *config.Config) (repository.PacketHandler, error) {
	psTLSPortInt, err := strconv.Atoi(cfg.PSTLSPort)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ps tls port: %w", err)
	}

	psQUICPortInt, err := strconv.Atoi(cfg.PSQUICPort)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ps quic port: %w", err)
	}

	cmsTLSPortInt, err := strconv.Atoi(cfg.CMSTLSPort)
	if err != nil {
		return nil, fmt.Errorf("failed to convert cms tls port: %w", err)
	}

	cmsQUICPortInt, err := strconv.Atoi(cfg.CMSQUICPort)
	if err != nil {
		return nil, fmt.Errorf("failed to convert cms quic port: %w", err)
	}

	return &packetHandler{
		psFQDN:      cfg.PSFQDN,
		psTLSPort:   psTLSPortInt,
		psQUICPort:  psQUICPortInt,
		cmsFQDN:     cfg.CMSFQDN,
		cmsTLSPort:  cmsTLSPortInt,
		cmsQUICPort: cmsQUICPortInt,
	}, nil
}

func (ph *packetHandler) Unmarshal(ctx context.Context, buf []byte, length int) (*entity.LoginRequest, bool, error) {
	if length < entity.BaseLoginRequestLength {
		return nil, false, fmt.Errorf("invalid login request packet length")
	}

	shouldReadPacket := false

	if buf[5] == uint8(entity.TypeClassLoginRequest) {
		req := entity.LoginRequest{
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
			AuthType:          entity.TypeAuthClass(binary.BigEndian.Uint16(buf[32:34])),
			DeviceIDLength:    0,
			PasswordLength:    0,
			CertificateLength: 0,
			DeviceID:          []byte{},
			Password:          []byte{},
			Certificate:       []byte{},
		}

		copy(req.BaseHeader.ID, buf[16:32])

		DeviceIDLength := binary.BigEndian.Uint16(buf[34:36])
		PasswordLength := binary.BigEndian.Uint16(buf[36:38])
		CertificateLength := binary.BigEndian.Uint16(buf[38:40])

		switch req.AuthType {
		case entity.TypeDeviceIDPassword:
			if uint16(length) != binary.BigEndian.Uint16(buf[12:14]) {
				return nil, false, fmt.Errorf("invalid login request packet length")
			}

			req.DeviceIDLength = DeviceIDLength
			req.PasswordLength = PasswordLength

			deviceIDOffset := uint16(entity.BaseLoginRequestLength)
			passwordOffset := deviceIDOffset + DeviceIDLength

			req.DeviceID = buf[deviceIDOffset : deviceIDOffset+DeviceIDLength]
			req.Password = buf[passwordOffset : passwordOffset+PasswordLength]

		case entity.TypeDigitalAuthentication:
			req.CertificateLength = CertificateLength

			certificateOffset := uint16(entity.BaseLoginRequestLength)

			req.Certificate = append(req.Certificate, buf[certificateOffset:length]...)

			if int(req.CertificateLength) > len(req.Certificate) {
				shouldReadPacket = true
			} else if int(req.CertificateLength) < len(req.Certificate) {
				return nil, false, fmt.Errorf("invalid request due to large certificate")
			}

		case entity.TypeSingleSignOn:

		default:
			return nil, false, fmt.Errorf("unexpected auth type")
		}

		return &req, shouldReadPacket, nil
	}

	return nil, false, fmt.Errorf("invalid packet type of login request")
}

func (ph *packetHandler) ReCreateLoginRequest(ctx context.Context, req *entity.LoginRequest, buf []byte, length int) *entity.LoginRequest {
	req.Certificate = append(req.Certificate, buf[:length]...)

	return req
}

func (ph *packetHandler) GenerateLoginResponse() *entity.LoginResponse {
	return &entity.LoginResponse{
		BaseHeader: entity.BaseHeader{
			TransactionID:  0,
			Version:        0,
			Type:           uint8(entity.TypeClassLoginResponse),
			Status:         0,
			Count:          0,
			SequenceNumber: 0,
			MessageLength:  entity.BaseLoginResponseLength,
			Token:          0,
			NextOpt:        0,
			ID:             make([]byte, entity.IDlen),
		},
		SecretCommonValue: make([]byte, entity.SecretCommonValueSize),
		PSFQDNLength:      0,
		PSPort:            0,
		CMSFQDNLength:     0,
		CMSPort:           0,
		ExpireDate: entity.ExpireDate{
			Year:  0,
			Month: 0,
			Day:   0,
		},
		AccessTokenLength: 0,
		Padding:           0,
		PSFQDN:            []byte{},
		CMSFQDN:           []byte{},
		AccessToken:       []byte{},
	}
}

func (ph *packetHandler) SerializeBaseHeader(res *entity.LoginResponse, prevBaseHeader entity.BaseHeader, st entity.StatusClass, nodeID []byte) {
	res.SerializeBaseHeader(prevBaseHeader, entity.TypeClassLoginResponse, st, nodeID)
}

func (ph *packetHandler) SerializeSecretCommonValue(res *entity.LoginResponse) {
	res.SerializeSecretCommonValue()
}

func (ph *packetHandler) SerializePSInformation(res *entity.LoginResponse, isQUIC bool) {
	if isQUIC {
		res.SerializePSInformation(ph.psFQDN, ph.psQUICPort)
	} else {
		res.SerializePSInformation(ph.psFQDN, ph.psTLSPort)
	}
}

func (ph *packetHandler) SerializeCMSInformation(res *entity.LoginResponse, isQUIC bool) {
	if isQUIC {
		res.SerializeCMSInformation(ph.cmsFQDN, ph.cmsQUICPort)
	} else {
		res.SerializeCMSInformation(ph.cmsFQDN, ph.cmsTLSPort)
	}
}

func (ph *packetHandler) SerializeAccessToken(res *entity.LoginResponse, token string) {
	res.SerializeAccessToken(token)
}

func (ph *packetHandler) ChangeStatus(res *entity.LoginResponse, st entity.StatusClass) {
	res.ChangeStatus(st)
}
