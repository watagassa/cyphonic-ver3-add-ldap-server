package entity

import (
	"encoding/binary"
	"errors"
	"time"
)

// LoginResponse is Login Response of a CYPHONIC packet.
// CipherType sets the algorithm used when encrypting and decrypting.
type LoginResponse struct {
	BaseHeader        BaseHeader
	SecretCommonValue []byte
	PSFQDNLength      uint16
	PSPort            uint16
	CMSFQDNLength     uint16
	CMSPort           uint16
	ExpireDate        ExpireDate
	AccessTokenLength uint16
	Padding           uint16
	PSFQDN            []byte
	CMSFQDN           []byte
	AccessToken       []byte
}

// ExpireDate handles Common Key expiration.
type ExpireDate struct {
	Year  uint16
	Month uint8
	Day   uint8
}

const (
	// YearSize is length of year (bytes).
	YearSize = 2

	// MonthSize is length of month (bytes).
	MonthSize = 1

	// DaySize is length of day (bytes).
	DaySize = 1
)

// Payload length in LoginResponse.
const (
	// SecretCommonValue lengths (bytes).
	SecretCommonValueSize = 16

	// PS's FQDN size lengths (bytes).
	PSFQDNLengthSize = 2

	// Ps's Port length (bytes).
	PSPortSize = 2

	// CMS's FQDN size lengths (bytes).
	CMSFQDNLengthSize = 2

	// CMS's Port lengths (bytes).
	CMSPortSize = 2

	// ExpireDate length (bytes).
	ExpireDateSize = YearSize + MonthSize + DaySize

	// Access Token size lengths (bytes).
	AccessTokenLengthSize = 2

	// PaddingSize lengths (bytes).
	PaddingSize = 2
)

// BaseLoginResponseLength is length of (base) body part of LoginResponse.
// Also, since the length of FQDN and CommonKey is variable length, when
// marshaling, calculate and add.
const BaseLoginResponseLength = BaseHeaderLen +
	SecretCommonValueSize +
	PSFQDNLengthSize + PSPortSize + CMSFQDNLengthSize + CMSPortSize +
	ExpireDateSize + AccessTokenLengthSize + PaddingSize

func (response *LoginResponse) SerializeSecretCommonValue() {
	response.SecretCommonValue = []byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
}

func (response *LoginResponse) SerializePSInformation(fqdn string, port int) {
	response.PSFQDN = []byte(fqdn)
	response.PSFQDNLength = uint16(len(fqdn))
	response.PSPort = uint16(port)
}

func (response *LoginResponse) SerializeCMSInformation(fqdn string, port int) {
	response.CMSFQDN = []byte(fqdn)
	response.CMSFQDNLength = uint16(len(fqdn))
	response.CMSPort = uint16(port)
}

func (response *LoginResponse) SerializeAccessToken(accessToken string) {
	expireDate := time.Now().AddDate(0, 1, 0)

	response.ExpireDate.Year = uint16(expireDate.Year())
	response.ExpireDate.Month = uint8(expireDate.Month())
	response.ExpireDate.Day = uint8(expireDate.Day())

	response.AccessToken = []byte(accessToken)
	response.AccessTokenLength = uint16(len(accessToken))
}

func (response *LoginResponse) SerializeBaseHeader(prevBaseHeader BaseHeader, typ TypeClass, st StatusClass, nodeID []byte) {
	response.BaseHeader.serializeTransactionID(prevBaseHeader)
	response.BaseHeader.serializeVersion()
	response.BaseHeader.serializeType(typ)
	response.BaseHeader.serializeStatus(st)
	response.BaseHeader.serializeCount()
	response.BaseHeader.serializeSequenceNumber(prevBaseHeader)
	response.BaseHeader.serializeMessageLength(uint16(BaseLoginResponseLength + response.PSFQDNLength + response.CMSFQDNLength + response.AccessTokenLength))
	response.BaseHeader.serializeToken()
	response.BaseHeader.serializeNextOpt()
	response.BaseHeader.serializeID(nodeID)
}

func (response *LoginResponse) ChangeStatus(st StatusClass) {
	response.BaseHeader.serializeStatus(st)
}

func (response *LoginResponse) ConvertBytes() ([]byte, error) {
	buf := make([]byte, BaseLoginResponseLength, BaseLoginResponseLength+response.PSFQDNLength+response.CMSFQDNLength+response.AccessTokenLength)

	// Base Header
	binary.BigEndian.PutUint32(buf[0:4], response.BaseHeader.TransactionID)
	buf[4] = response.BaseHeader.Version
	buf[5] = response.BaseHeader.Type
	buf[6] = response.BaseHeader.Status
	buf[7] = response.BaseHeader.Count
	binary.BigEndian.PutUint32(buf[8:12], response.BaseHeader.SequenceNumber)
	binary.BigEndian.PutUint16(buf[12:14], response.BaseHeader.MessageLength)
	buf[14] = response.BaseHeader.Token
	buf[15] = response.BaseHeader.NextOpt
	binary.BigEndian.PutUint64(buf[16:24], binary.BigEndian.Uint64(response.BaseHeader.ID[0:8]))
	binary.BigEndian.PutUint64(buf[24:32], binary.BigEndian.Uint64(response.BaseHeader.ID[8:16]))

	// Login Response Body
	binary.BigEndian.PutUint64(buf[32:40], binary.BigEndian.Uint64(response.SecretCommonValue[0:8]))
	binary.BigEndian.PutUint64(buf[40:48], binary.BigEndian.Uint64(response.SecretCommonValue[8:16]))
	binary.BigEndian.PutUint16(buf[48:50], response.PSFQDNLength)
	binary.BigEndian.PutUint16(buf[50:52], response.PSPort)
	binary.BigEndian.PutUint16(buf[52:54], response.CMSFQDNLength)
	binary.BigEndian.PutUint16(buf[54:56], response.CMSPort)
	binary.BigEndian.PutUint16(buf[56:58], response.ExpireDate.Year)
	buf[58] = response.ExpireDate.Month
	buf[59] = response.ExpireDate.Day
	binary.BigEndian.PutUint16(buf[60:62], response.AccessTokenLength)
	binary.BigEndian.PutUint16(buf[62:64], response.Padding)

	if len(response.PSFQDN) != int(response.PSFQDNLength) {
		return nil, errors.New("ps fqdn does not match fqdn length")
	}

	buf = append(buf, response.PSFQDN...)

	if len(response.CMSFQDN) != int(response.CMSFQDNLength) {
		return nil, errors.New("cms fqdn does not match fqdn length")
	}

	buf = append(buf, response.CMSFQDN...)

	if len(response.AccessToken) != int(response.AccessTokenLength) {
		return nil, errors.New("access token does not match access token length")
	}

	buf = append(buf, response.AccessToken...)

	if len(buf) != int(response.BaseHeader.MessageLength) {
		return nil, errors.New("message does not match MessageLength")
	}

	return buf, nil
}
