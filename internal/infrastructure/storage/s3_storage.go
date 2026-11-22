package storage

import (
	// "github.com/aws/aws-sdk-go-v2/aws"
	"context"
	"ecommerce/internal/usecase/interfaces"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// "github.com/aws/smithy-go/context"
)

type S3Storage struct {
	client *s3.Client
	bucket string
	region string
}
type S3Config struct {
	Bucket string
	Region string
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	client := s3.NewFromConfig(awsCfg)
	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
	}, nil
}
func (s *S3Storage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload object to S3: %w", err)
	}
	objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, objectKey)
	return objectURL, nil
}
func (s *S3Storage) Delete(ctx context.Context, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(objectKey)})
	if err != nil {
		return fmt.Errorf("delete object from S3: %w", err)
	}
	return nil
}
func (s *S3Storage) objectKeyFromURL(imageURL string) (string, error) {
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return "", fmt.Errorf("parse image URL: %w", err)
	}
	expectedHost := fmt.Sprintf(
		"%s.s3.%s.amazonaws.com",
		s.bucket,
		s.region,
	)
	if parsedURL.Host != expectedHost {
		return "", fmt.Errorf("invalid storage URL")
	}
	objectKey := strings.TrimPrefix(parsedURL.Path, "/")
	if objectKey == "" {
		return "", fmt.Errorf("invalid object key")
	}
	return objectKey, nil
}

func (s *S3Storage) DeleteByURL(ctx context.Context, imageURL string) error {
	objectKey, err := s.objectKeyFromURL(imageURL)
	if err != nil {
		return err
	}
	return s.Delete(ctx, objectKey)
}

func (s *S3Storage) DownloadByURL(ctx context.Context, imageURL string) (io.ReadCloser, string, error) {
	objectKey, err := s.objectKeyFromURL(imageURL)
	if err != nil {
		return nil, "", err
	}
	object, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, "", fmt.Errorf("download object from S3: %w", err)
	}
	return object.Body, aws.ToString(object.ContentType), nil
}

var _ interfaces.ObjectStorage = (*S3Storage)(nil)
