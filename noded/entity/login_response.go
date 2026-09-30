package entity

import (
	"encoding/binary"
	"errors"
	"fmt"
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
const BaseLoginResponseLength = BaseHeaderLength +
	SecretCommonValueSize +
	PSFQDNLengthSize + PSPortSize + CMSFQDNLengthSize + CMSPortSize +
	ExpireDateSize + AccessTokenLengthSize + PaddingSize

var ErrInvalidBufferLength = errors.New("invalid buffer length")

func UnmarshalLoginResponse(buffer []byte) (*LoginResponse, error) {
	if len(buffer) < BaseLoginResponseLength {
		return nil, fmt.Errorf("%w: %d", ErrInvalidBufferLength, len(buffer))
	}

	baseHeader := UnmarshalBaseHeader(buffer[:BaseHeaderLength])
	psFQDNLength := binary.BigEndian.Uint16(buffer[BaseHeaderLength+SecretCommonValueSize : BaseHeaderLength+SecretCommonValueSize+PSFQDNLengthSize])
	psFQDNStart := BaseHeaderLength + SecretCommonValueSize + PSFQDNLengthSize + PSPortSize +
		CMSFQDNLengthSize + CMSPortSize + ExpireDateSize + AccessTokenLengthSize + PaddingSize
	psFQDNEnd := psFQDNStart + int(psFQDNLength)
	cmsFQDNLength := binary.BigEndian.Uint16(buffer[BaseHeaderLength+SecretCommonValueSize+
		PSFQDNLengthSize+PSPortSize : BaseHeaderLength+SecretCommonValueSize+PSFQDNLengthSize+
		PSPortSize+CMSFQDNLengthSize])
	cmsFQDNEnd := psFQDNEnd + int(cmsFQDNLength)
	accessTokenLength := binary.BigEndian.Uint16(buffer[BaseHeaderLength+SecretCommonValueSize+
		PSFQDNLengthSize+PSPortSize+CMSFQDNLengthSize+CMSPortSize+ExpireDateSize : BaseHeaderLength+SecretCommonValueSize+PSFQDNLengthSize+PSPortSize+CMSFQDNLengthSize+
		CMSPortSize+ExpireDateSize+AccessTokenLengthSize])
	accessTokenEnd := cmsFQDNEnd + int(accessTokenLength)

	return &LoginResponse{
		BaseHeader:        *baseHeader,
		SecretCommonValue: buffer[BaseHeaderLength : BaseHeaderLength+SecretCommonValueSize],
		PSFQDNLength:      psFQDNLength,
		PSPort: binary.BigEndian.Uint16(buffer[BaseHeaderLength+SecretCommonValueSize+
			PSFQDNLengthSize : BaseHeaderLength+SecretCommonValueSize+PSFQDNLengthSize+PSPortSize]),
		CMSFQDNLength: cmsFQDNLength,
		CMSPort: binary.BigEndian.Uint16(buffer[BaseHeaderLength+SecretCommonValueSize+
			PSFQDNLengthSize+PSPortSize+CMSFQDNLengthSize : BaseHeaderLength+SecretCommonValueSize+
			PSFQDNLengthSize+PSPortSize+CMSFQDNLengthSize+CMSPortSize]),
		ExpireDate: ExpireDate{
			Year: binary.BigEndian.Uint16(buffer[BaseHeaderLength+SecretCommonValueSize+
				PSFQDNLengthSize+PSPortSize+CMSFQDNLengthSize+CMSPortSize : BaseHeaderLength+
				SecretCommonValueSize+PSFQDNLengthSize+PSPortSize+CMSFQDNLengthSize+CMSPortSize+
				YearSize]),
			Month: buffer[BaseHeaderLength+SecretCommonValueSize+PSFQDNLengthSize+PSPortSize+
				CMSFQDNLengthSize+CMSPortSize+YearSize],
			Day: buffer[BaseHeaderLength+SecretCommonValueSize+PSFQDNLengthSize+PSPortSize+
				CMSFQDNLengthSize+CMSPortSize+YearSize+MonthSize],
		},
		AccessTokenLength: accessTokenLength,
		Padding:           binary.BigEndian.Uint16(buffer[accessTokenEnd : accessTokenEnd+PaddingSize]),
		PSFQDN:            buffer[psFQDNStart:psFQDNEnd],
		CMSFQDN:           buffer[psFQDNEnd:cmsFQDNEnd],
		AccessToken:       buffer[cmsFQDNEnd:accessTokenEnd],
	}, nil
}
