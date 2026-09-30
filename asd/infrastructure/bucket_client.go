//nolint:exhaustruct // in this file, aws-sdk-go package has many parameters and unnecessary key-values reduce readability.
package infrastructure

import (
	context "context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Pluslab/cyphonic/asd/usecase/repository"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type bucketClient struct {
	S3 *s3.S3
}

func NewBucketClient(accessKey, secretKey, region, endPoint string) (repository.BucketClient, error) {
	bucketSession, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("can't create new bucket session: %w", err)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{
		Transport: tr,
	}

	cfg := aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Region:           aws.String(region),
		Endpoint:         aws.String(endPoint),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       httpClient,
	}

	s3Client := s3.New(bucketSession, &cfg)

	return &bucketClient{
		S3: s3Client,
	}, nil
}

const timeout = 30 * time.Second

func (c *bucketClient) Read(bucketName, objectKey string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	objectRequest := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	resultCh := make(chan *s3.GetObjectOutput, 1)
	errCh := make(chan error, 1)

	go func() {
		object, err := c.S3.GetObjectWithContext(ctx, objectRequest)
		if err != nil {
			errCh <- fmt.Errorf("can't get object: %w", err)

			return
		}
		resultCh <- object
	}()

	select {
	case object := <-resultCh:
		defer object.Body.Close()

		bodyBytes, err := io.ReadAll(object.Body)
		if err != nil {
			return nil, fmt.Errorf("can't read object body")
		}

		return bodyBytes, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("object retrieval took too long")
	}
}
