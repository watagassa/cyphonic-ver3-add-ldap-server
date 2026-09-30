// Package certificate contains the certficate of TLS connetction.
package certificate

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

// GenerateCYPClientCert return cyphonic client certificate.
func GenerateCYPClientCert(sNumber int, cName, nEmail, fqdn string) (ccertBytes []byte, err error) {

	as, asPrivKey, err := GetASCA()
	if err != nil {
		return nil, err
	}

	cPubKey, err := GetClientPubKey(cName)
	if err != nil {
		return nil, err
	}

	cert := SetUpCYPClientCert(sNumber, cName, nEmail, fqdn)

	ccertBytes, err = x509.CreateCertificate(rand.Reader, cert, as, cPubKey, asPrivKey)

	if err != nil {
		return nil, err
	}

	err = CreateClientCert(ccertBytes, cName)
	if err != nil {
		return nil, err
	}

	return ccertBytes, nil
}

// SetUpCYPClientCert is setting up cyphonic client certificate struct.
func SetUpCYPClientCert(serialNum int, commonName, nodeEmail, fqdn string) (cypClient *x509.Certificate) {
	var expandYears int = 10

	// set up our cyphonic client certificate
	cypClient = &x509.Certificate{
		SerialNumber: big.NewInt(int64(serialNum)),
		Subject: pkix.Name{
			Organization:       []string{"Pluslab"},
			OrganizationalUnit: []string{"Pluslab"},
			Country:            []string{"JP"},
			Province:           []string{"Aichi"},
			Locality:           []string{"Toyota"},
			StreetAddress:      []string{""},
			PostalCode:         []string{""},
			CommonName:         fqdn,
		},
		NotBefore:      time.Now(),
		NotAfter:       time.Now().AddDate(expandYears, 0, 0), // 10 years
		EmailAddresses: []string{nodeEmail},
		KeyUsage:       x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageContentCommitment,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames:       []string{fqdn},
	}

	return cypClient
}
