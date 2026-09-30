package entity

import (
	"bytes"
	"crypto/sha512"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
)

// LoginRequest is Login Request of a CYPHONIC packet.
// AuthType sets the authentication method for authentication service.
// ExecMode handles whether you are a user or an administrator.
type LoginRequest struct {
	BaseHeader        BaseHeader
	AuthType          TypeAuthClass
	DeviceIDLength    uint16
	PasswordLength    uint16
	CertificateLength uint16
	DeviceID          DeviceID
	Password          Password
	Certificate       Certificate
}

// Payload length in LoginRequest.
const (
	// AuthTypeSize lengths (bytes).
	AuthTypeSize = 2

	// DeviceIDLengthSize lengths (bytes).
	DeviceIDLengthSize = 2

	// PasswordLengthSize lengths (bytes).
	PasswordLengthSize = 2

	// CertificateLengthSize lengths (bytes).
	CertificateLengthSize = 2
)

var (
	ErrDeviceIDEmpty                  = errors.New("deviceID is empty")
	ErrPasswordEmpty                  = errors.New("password is empty")
	ErrFailedToDecodeBlockCertificate = errors.New("failed to decode block certificate")
)

// BaseLoginRequestLength is length of (base) body part of LoginRequest.
// Also, since the length of DeviceID and Password or Certificate is variable length, when
// marshaling, calculate and add.
const BaseLoginRequestLength = BaseHeaderLength + AuthTypeSize + DeviceIDLengthSize + PasswordLengthSize + CertificateLengthSize

// DeviceID is the deviceID for logging in.
type DeviceID []byte

// Password is the password to login.
type Password []byte

// Certificate is the certificate to login.
type Certificate []byte

// TypeAuthClass defines the class associated with Auth type in CYPHONIC packets.
type TypeAuthClass uint16

// TypeAuthClass known values.
const (
	TypeDeviceIDPassword      TypeAuthClass = 1
	TypeDigitalAuthentication TypeAuthClass = 2
	TypeSingleSignOn          TypeAuthClass = 3
)

// GetDigitalCertificate get digital certificate.
func (lreq LoginRequest) GetDigitalCertificate() (*x509.Certificate, error) {
	block, _ := pem.Decode(lreq.Certificate)
	if block == nil {
		return nil, ErrFailedToDecodeBlockCertificate
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed parse certificate: %w", err)
	}

	return cert, nil
}

func GenerateLoginRequest(deviceID, password string) (*LoginRequest, error) {
	if len(deviceID) == 0 {
		return nil, ErrDeviceIDEmpty
	} else if len(password) == 0 {
		return nil, ErrPasswordEmpty
	}

	baseHeader, err := GenerateBaseHeader(TypeClassLoginRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base header: %w", err)
	}

	b := []byte(password)
	sha512 := sha512.Sum512(b)

	return &LoginRequest{
		BaseHeader:        *baseHeader,
		AuthType:          TypeDeviceIDPassword,
		DeviceIDLength:    uint16(len(deviceID)),
		PasswordLength:    uint16(len(sha512)),
		CertificateLength: 0,
		DeviceID:          []byte(deviceID),
		Password:          sha512[:],
		Certificate:       nil,
	}, nil
}

func (lreq LoginRequest) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	lreq.BaseHeader.MessageLength = BaseLoginRequestLength + lreq.DeviceIDLength + lreq.PasswordLength + lreq.CertificateLength

	baseHeaderBuf := lreq.BaseHeader.All()
	if _, err := buffer.Write(baseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	fields := []interface{}{
		lreq.AuthType,
		lreq.DeviceIDLength,
		lreq.PasswordLength,
		lreq.CertificateLength,
		lreq.DeviceID,
		lreq.Password,
		lreq.Certificate,
	}
	for _, field := range fields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to write field: %w", err)
		}
	}

	return buffer.Bytes(), nil
}
