// Package certificate contains the certficate of TLS connetction.
package certificate

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/Pluslab/cyphonic/ops/logger"
)

// GetRootCA read ca cert and caPrivKey
func GetRootCA() (ca *x509.Certificate, caPrivKey *rsa.PrivateKey, err error) {

	caPrivKey, err = ReadCAPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	ca, err = ReadCACertificate()
	if err != nil {
		return nil, nil, err
	}

	return ca, caPrivKey, nil
}

// ReadCAPrivateKey read ca PrivKey
func ReadCAPrivateKey() (caPrivKey *rsa.PrivateKey, err error) {
	PriKey, err := os.ReadFile("ca_private_key.pem")
	if err != nil {
		logger.LogErr("RootCA: Could not load ca privtekey", "error", err)
		return nil, err
	}

	PrivKeyBlock, _ := pem.Decode(PriKey)
	if PrivKeyBlock == nil {
		err = errors.New("invalid private key data")
		logger.LogErr("RootCA: private key data faild", "error", err)
		return nil, err
	}

	if PrivKeyBlock.Type == "RSA PRIVATE KEY" {
		caPrivKey, err = x509.ParsePKCS1PrivateKey(PrivKeyBlock.Bytes)
		if err != nil {
			logger.LogErr("RootCA: Failed to parse private key", "error", err)
			return nil, err
		}
	} else {
		err = errors.New("invalid private key type")
		logger.LogErr("RootCA: Error private key type", "error", err)
		return nil, err
	}

	caPrivKey.Precompute()

	if err := caPrivKey.Validate(); err != nil {
		logger.LogErr("RootCA: Error Integrity of　private key", "error", err)
		return nil, err
	}

	return caPrivKey, nil
}

// ReadCACertificate read ca Certificate
func ReadCACertificate() (ca *x509.Certificate, err error) {
	Cert, err := os.ReadFile("ca.pem")
	if err != nil {
		logger.LogErr("RootCA: Could not load ca certificate", "error", err)
		return nil, err
	}
	CAPool := x509.NewCertPool()
	if ok := CAPool.AppendCertsFromPEM(Cert); !ok {
		logger.LogErr("RootCA: The certificate is not correct.")
		return nil, err
	}

	CertBlock, _ := pem.Decode(Cert)
	if CertBlock == nil {
		err = errors.New("invalid private key data")
		logger.LogErr("RootCA: ca certificate data faild", "error", err)
		return nil, err
	}
	ca, err = x509.ParseCertificate(CertBlock.Bytes)
	if err != nil {
		logger.LogErr("RootCA: failed to parse certificate", "error", err)
		return nil, err
	}

	return ca, nil
}
