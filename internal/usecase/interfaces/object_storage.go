package interfaces

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, objectKey string) error
	DeleteByURL(ctx context.Context, imageURL string) error
}
