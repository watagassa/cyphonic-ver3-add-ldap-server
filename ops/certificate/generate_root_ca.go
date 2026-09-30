// Package certificate contains the certficate of TLS connetction.
package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"github.com/Pluslab/cyphonic/ops/logger"
)

// GenerateRootCA creates RootCA.
func GenerateRootCA() (ca *x509.Certificate, caPrivKey *rsa.PrivateKey, err error) {
	ca = SetUpRootCA("CYPRootCA")

	// create our private and public key
	caPrivKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	if err = CreateCA(ca, caPrivKey); err != nil {
		return nil, nil, err
	}

	if err = CreateCAPrivateKeyPEM(ca, caPrivKey); err != nil {
		return nil, nil, err
	}

	return ca, caPrivKey, nil
}

// SetUpRootCA is setting up RootCA.
func SetUpRootCA(commonName string) (ca *x509.Certificate) {
	var serialNum int64 = 2021
	var expandYears int = 10

	// set up our CA certificate
	ca = &x509.Certificate{
		SerialNumber: big.NewInt(serialNum),
		Subject: pkix.Name{
			Organization:       []string{"Pluslab"},
			OrganizationalUnit: []string{"Pluslab"},
			Country:            []string{"JP"},
			Province:           []string{"Aichi"},
			Locality:           []string{"Toyota"},
			StreetAddress:      []string{""},
			PostalCode:         []string{""},
			CommonName:         commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(expandYears, 0, 0), // 10 years
		IsCA:                  true,
		EmailAddresses:        []string{"yoshikawataiki@pluslab.org"},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	return ca
}

// CreateCA creates RootCA.
func CreateCA(ca *x509.Certificate, caPrivKey *rsa.PrivateKey) (err error) {
	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		logger.LogErr("RootCA: x.509 certificate creation failed", "error", err)
		return err
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		logger.LogErr("RootCA: Failed to create Buffer string", "error", err)
		return err
	}

	f, err := os.Create("ca.pem")
	if err != nil {
		logger.LogErr("RootCA: Failed to create ca.pem", "error", err)
		return err
	}

	err = pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes})
	if err != nil {
		logger.LogErr("RootCA: ca.pem encoding failure", "error", err)
		return err
	}

	err = f.Close()
	if err != nil {
		logger.LogErr("RootCA: Failed to close ca.pem", "error", err)
		return err
	}

	return nil
}

// CreateCAPrivateKeyPEM creates Private Key.
func CreateCAPrivateKeyPEM(ca *x509.Certificate, caPrivKey *rsa.PrivateKey) (err error) {
	f, err := os.Create("ca_private_key.pem")
	if err != nil {
		logger.LogErr("RootCA-private: Failed to create ca_private_key.pem", "error", err)
		return err
	}

	// pem encode
	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	if err != nil {
		logger.LogErr("RootCA-private: Failed to create Buffer string", "error", err)
		return err
	}

	err = pem.Encode(f, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey)})
	if err != nil {
		logger.LogErr("RootCA-private: ca_private_key.pem encoding failure", "error", err)
		return err
	}

	err = f.Close()
	if err != nil {
		logger.LogErr("RootCA-private: Failed to close ca_private_key.pem", "error", err)
		return err
	}

	return nil
}
