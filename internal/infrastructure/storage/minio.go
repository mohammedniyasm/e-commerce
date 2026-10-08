package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

func NewMinIOStorage(
	cfg MinIOConfig,
) (*MinIOStorage, error) {

	client, err := minio.New(
		cfg.Endpoint,
		&minio.Options{
			Creds: credentials.NewStaticV4(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
			Secure: cfg.UseSSL,
		},
	)
	if err != nil {
		return nil, err
	}

	storage := &MinIOStorage{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
		useSSL:   cfg.UseSSL,
	}

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(
			ctx,
			cfg.Bucket,
			minio.MakeBucketOptions{},
		); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

func (s *MinIOStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectKey,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", err
	}

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}

	url := fmt.Sprintf(
		"%s://%s/%s/%s",
		scheme,
		s.endpoint,
		s.bucket,
		objectKey,
	)

	return url, nil
}

func (s *MinIOStorage) Delete(
	ctx context.Context,
	objectKey string,
) error {
	return s.client.RemoveObject(
		ctx,
		s.bucket,
		objectKey,
		minio.RemoveObjectOptions{},
	)
}
func (s *MinIOStorage) DeleteByURL(
	ctx context.Context,
	imageURL string,
) error {

	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return err
	}

	// Expected URL:
	// http://localhost:9000/ecommerce/products/1/images/abc123

	path := strings.TrimPrefix(parsedURL.Path, "/")

	prefix := s.bucket + "/"

	if !strings.HasPrefix(path, prefix) {                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               
		return fmt.Errorf("invalid storage URL")
	}

	objectKey := strings.TrimPrefix(path, prefix)

	if objectKey == "" {
		return fmt.Errorf("invalid object key")
	}

	return s.Delete(ctx, objectKey)
}
