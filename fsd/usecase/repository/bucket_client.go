//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/fsd/$GOPACKAGE
package repository

type BucketClient interface {
	Read(bucketName, objectKey string) ([]byte, error)
}
