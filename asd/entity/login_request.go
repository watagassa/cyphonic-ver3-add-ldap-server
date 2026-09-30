package entity

import (
	"crypto/x509"
	"encoding/pem"
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

// ApplicationID is a CYPHONIC Application ID, a slice of bytes.
// length of the byte slice: a 16-byte slice.
type ApplicationID []byte

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

// BaseLoginRequestLength is length of (base) body part of LoginRequest.
// Also, since the length of DeviceID and Password or Certificate is variable length, when
// marshaling, calculate and add.
const BaseLoginRequestLength = BaseHeaderLen + AuthTypeSize + DeviceIDLengthSize + PasswordLengthSize + CertificateLengthSize

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
		return nil, fmt.Errorf("failed to decode block certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed parse certificate: %w", err)
	}

	return cert, nil
}
