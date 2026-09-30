package infrastructure

import (
	"fmt"
	"os"
	"sync"

	entity "github.com/Pluslab/cyphonic/cmsd/entity"
	"github.com/Pluslab/cyphonic/cmsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/cmsd/usecase/repository"
)

// certificate defines the certificate entity.
type certificate struct {
	Certificate []byte
}

// CertificateStore acts as an kvs-based internal cache.
type CertificateStoreCache struct {
	certificateStores map[string]*certificate
	rootCertName      string
	mu                sync.Mutex
}

func NewCertificateStore(fqdn, rootCertName, accessKey, secretKey, region, endPoint, bucketName, certificateLocation string) (repository.CertificateStore, error) {
	certStore := &CertificateStoreCache{
		certificateStores: make(map[string]*certificate),
		rootCertName:      rootCertName,
		mu:                sync.Mutex{},
	}

	if err := certStore.Set(fqdn, rootCertName, accessKey, secretKey, region, endPoint, bucketName, certificateLocation); err != nil {
		return nil, fmt.Errorf("can't set certificate: %w", err)
	}

	return certStore, nil
}

func (s *CertificateStoreCache) Set(fqdn, rootCertName, accessKey, secretKey, region, endPoint, bucketName, certificateLocation string) error {
	var client repository.BucketClient

	var (
		rootCert          []byte
		serverCert        []byte
		serverCertPrivKey []byte
	)

	var err error

	serverCertName := fmt.Sprintf("%s.pem", fqdn)
	serverCertPrivKeyName := fmt.Sprintf("%s.key", fqdn)

	config.LogInfo(fmt.Sprintf("LocationType: %s - Bucket EndPoint: %s\n", certificateLocation, endPoint))

	switch certificateLocation {
	case entity.Local:
		rootCert, err = os.ReadFile("./" + rootCertName)
		if err != nil {
			return fmt.Errorf("can't get root certificate: %w", err)
		}

		s.set(entity.RootCert, rootCert)

		serverCert, err = os.ReadFile("./" + serverCertName)
		if err != nil {
			return fmt.Errorf("can't get server certificate: %w", err)
		}

		s.set(entity.ServerCert, serverCert)

		serverCertPrivKey, err = os.ReadFile("./" + serverCertPrivKeyName)
		if err != nil {
			return fmt.Errorf("can't get server certificate: %w", err)
		}

		s.set(entity.ServerCertPrivKey, serverCertPrivKey)

	case entity.Remote:
		client, err = NewBucketClient(accessKey, secretKey, region, endPoint)
		if err != nil {
			return fmt.Errorf("can't activate client bucket: %w", err)
		}

		rootCert, err = client.Read(bucketName, rootCertName)
		if err != nil {
			return fmt.Errorf("can't get root certificate: %w", err)
		}

		s.set(entity.RootCert, rootCert)

		serverCert, err = client.Read(bucketName, serverCertName)
		if err != nil {
			return fmt.Errorf("can't get server certificate: %w", err)
		}

		s.set(entity.ServerCert, serverCert)

		serverCertPrivKey, err = client.Read(bucketName, serverCertPrivKeyName)
		if err != nil {
			return fmt.Errorf("can't get server certificate private key: %w", err)
		}

		s.set(entity.ServerCertPrivKey, serverCertPrivKey)

	default:
		return fmt.Errorf("certificate location unknown")
	}

	return nil
}

func (s *CertificateStoreCache) set(key string, value []byte) {
	s.mu.Lock()
	if _, ok := s.certificateStores[key]; !ok {
		s.certificateStores[key] = &certificate{
			Certificate: value,
		}
	}
	s.mu.Unlock()
}

func (s *CertificateStoreCache) Get(certificateType string) []byte {
	cert := s.get(certificateType)

	return cert
}

func (s *CertificateStoreCache) get(key string) []byte {
	var value []byte

	s.mu.Lock()
	if v, ok := s.certificateStores[key]; ok {
		value = v.Certificate
	}
	s.mu.Unlock()

	return value
}

func (s *CertificateStoreCache) GetRootCertificateName() string {
	rootCertName := s.getRootCertificateName()

	return rootCertName
}

func (s *CertificateStoreCache) getRootCertificateName() string {
	s.mu.Lock()
	rootCertName := s.rootCertName
	s.mu.Unlock()

	return rootCertName
}
