// Package certificate contains the certficate of TLS connetction.
package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"github.com/Pluslab/cyphonic/ops/logger"
)

func GenerateServerCert(commonName string) (tlsConf *tls.Config, err error) {

	// set up RootCA
	ca, caPrivKey, err := GetRootCA()
	if err != nil {
		return nil, err
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	cert := SetUpServerCA(commonName)

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	certPEM, err := CreateCert(certBytes, commonName)
	if err != nil {
		return nil, err
	}

	tlsConf, err = CreateCertPrivate(certPrivKey, certPEM, commonName)
	if err != nil {
		return nil, err
	}

	return tlsConf, nil
}

// SetUpServerCA is setting up ServerCA.
func SetUpServerCA(commonName string) (ca *x509.Certificate) {
	// var serialNum int64 = 2021
	var expandYears int = 10

	serialNum, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		logger.LogErr("Failed to generate serial number", "error", err)
		return nil
	}

	// set up our CA certificate
	ca = &x509.Certificate{
		SerialNumber: serialNum,
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
		BasicConstraintsValid: true,
		IsCA:                  true,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(expandYears, 0, 0), // 10 years
		EmailAddresses:        []string{"yoshikawataiki@pluslab.org"},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		DNSNames:              []string{commonName},
	}
	return ca
}

// CreateCert creates server certificates.
func CreateCert(certBytes []byte, commonName string) (certPEM *bytes.Buffer, err error) {
	f, err := os.Create(commonName + ".pem")
	if err != nil {
		logger.LogErr("server: Failed to create FQDN.pem", "error", err)
		return nil, err
	}

	// pem encode
	certPEM = new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		logger.LogErr("server: Failed to create Buffer string", "error", err)
		return nil, err
	}

	err = pem.Encode(f, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		logger.LogErr("server: FQDN.pem encoding failure", "error", err)
		return nil, err
	}

	err = f.Close()
	if err != nil {
		logger.LogErr("server: Failed to close FQDN.pem", "error", err)
		return nil, err
	}

	return certPEM, nil
}

// CreateCertPrivate creates server private certificates.
func CreateCertPrivate(certPrivKey *rsa.PrivateKey, certPEM *bytes.Buffer, commonName string) (tlsConf *tls.Config, err error) {
	f, err := os.Create(commonName + ".key")
	if err != nil {
		logger.LogErr("server-private: Failed to create FQDN.pem", "error", err)
		return nil, err
	}

	// pem encode
	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	if err != nil {
		logger.LogErr("server-private: Failed to create Buffer string", "error", err)
		return nil, err
	}

	err = pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	if err != nil {
		logger.LogErr("server-private: FQDN.pem encoding failure", "error", err)
		return nil, err
	}

	err = f.Close()
	if err != nil {
		logger.LogErr("server-private: Failed to close FQDN.pem", "error", err)
		return nil, err
	}

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		logger.LogErr("server-private: Failed to parse the public and private key pair.", "error", err)
		return nil, err
	}

	tlsConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	return tlsConf, nil
}

// CreateClientCert creates client certificates.
func CreateClientCert(certBytes []byte, cName string) (err error) {
	f, err := os.Create(cName + "_cert.pem")
	if err != nil {
		logger.LogErr("client: Failed to create cName_cert.pem", "error", err)
		return err
	}

	err = pem.Encode(f, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		logger.LogErr("client: cName_cert.pem encoding failure", "error", err)
		return err
	}

	err = f.Close()
	if err != nil {
		logger.LogErr("client: Failed to close cName_cert.pem", "error", err)
		return err
	}

	return nil
}
