package certificate

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/Pluslab/cyphonic/ops/internal"
	"github.com/Pluslab/cyphonic/ops/logger"
)

// GetCert returns serverTLSConf to get server public and private key.
func GetCert() (tlsConf *tls.Config, err error) {
	cfg, err := internal.Get()

	serverCert, err := tls.LoadX509KeyPair("./"+cfg.FQDN+".pem", "./"+cfg.FQDN+".key")
	if err != nil {
		logger.LogErr("Failed to load the public and private key pair.", "error", err)
		return nil, err
	}

	CAPool := x509.NewCertPool()

	if caCert, err := os.ReadFile("ca.pem"); err != nil {
		logger.LogErr("Could not load ca certificate", "error", err)
	} else {
		if ok := CAPool.AppendCertsFromPEM(caCert); !ok {
			err = errors.New("the certificate is not correct")
			logger.LogErr("The certificate is not correct.", "error", err)
			return nil, err
		}
	}

	tlsConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		MinVersion:   tls.VersionTLS13,
		ClientCAs:    CAPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return tlsConf, nil
}

// GetASCA read as cert and asPrivKey
func GetASCA() (as *x509.Certificate, asPrivKey *rsa.PrivateKey, err error) {
	asPrivKey, err = ReadASPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	as, err = ReadASCertificate()
	if err != nil {
		return nil, nil, err
	}

	return as, asPrivKey, nil
}

// ReadASPrivateKey read as PrivKey
func ReadASPrivateKey() (asPrivKey *rsa.PrivateKey, err error) {
	PriKey, err := os.ReadFile(os.Getenv("FQDN") + ".key")
	if err != nil {
		logger.LogErr("RootCA: Could not load ca privtekey", "error", err)
		return nil, err
	}

	PrivKeyBlock, _ := pem.Decode(PriKey)
	if PrivKeyBlock == nil {
		err = errors.New("invalid private key data")
		logger.LogErr("AS: private key data faild", "error", err)
		return nil, err
	}

	if PrivKeyBlock.Type == "RSA PRIVATE KEY" {
		asPrivKey, err = x509.ParsePKCS1PrivateKey(PrivKeyBlock.Bytes)
		if err != nil {
			logger.LogErr("AS: Failed to parse private key", "error", err)
			return nil, err
		}
	} else {
		err = errors.New("invalid private key type")
		logger.LogErr("AS: Error private key type", "error", err)
		return nil, err
	}

	asPrivKey.Precompute()

	if err := asPrivKey.Validate(); err != nil {
		logger.LogErr("AS: Error Integrity of　private key", "error", err)
		return nil, err
	}

	return asPrivKey, nil
}

// ReadASCertificate read as Certificate
func ReadASCertificate() (as *x509.Certificate, err error) {
	Cert, err := os.ReadFile(os.Getenv("FQDN") + ".pem")
	if err != nil {
		logger.LogErr("AS: Could not load as certificate", "error", err)
		return nil, err
	}
	CAPool := x509.NewCertPool()
	if ok := CAPool.AppendCertsFromPEM(Cert); !ok {
		logger.LogErr("AS: The certificate is not correct.")
		return nil, err
	}

	CertBlock, _ := pem.Decode(Cert)
	if CertBlock == nil {
		err = errors.New("invalid private key data")
		logger.LogErr("AS: as certificate data faild", "error", err)
		return nil, err
	}
	as, err = x509.ParseCertificate(CertBlock.Bytes)
	if err != nil {
		logger.LogErr("AS: failed to parse certificate", "error", err)
		return nil, err
	}

	return as, nil
}

// GetClientPubKey read client PublicKey
func GetClientPubKey(cName string) (cPubKey *rsa.PublicKey, err error) {
	bytes, err := os.ReadFile(cName + ".pem")
	if err != nil {
		logger.LogErr("client-public: Could not load client public key", "error", err)
		return nil, err
	}

	PubKeyBlock, _ := pem.Decode(bytes)
	if PubKeyBlock == nil {
		err = errors.New("client-public: invalid public key data")
		logger.LogErr("client-public: public key data faild", "error", err)
		return nil, err
	}
	if PubKeyBlock.Type != "PUBLIC KEY" {
		err = errors.New("client-public: invalid public key type")
		logger.LogErr("client-public: Failed to parse public key", "error", err)
		return nil, err
	}
	cPubKey, err = x509.ParsePKCS1PublicKey(PubKeyBlock.Bytes)
	if err != nil {
		logger.LogErr("client-public: Failed to parse public key", "error", err)
		return nil, err
	}

	return cPubKey, nil
}
