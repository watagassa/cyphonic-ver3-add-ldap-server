//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/cmsd/$GOPACKAGE
package repository

type CertificateStore interface {
	Set(fqdn, rootCertName, accessKey, secretKey, region, endPoint, bucketName, certificateLocation string) error
	Get(certificateType string) []byte
	GetRootCertificateName() string
}
