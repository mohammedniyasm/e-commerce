package productimage

import (
	"errors"
	"testing"

	domainerrors "ecommerce/internal/domain/errors"
)

func TestValidateImage(t *testing.T) {
	tests := []struct {
		name        string
		size        int64
		contentType string
		wantError   error
	}{
		// Valid cases
		{
			name:        "valid JPEG image",
			size:        1024,
			contentType: "image/jpeg",
			wantError:   nil,
		},
		{
			name:        "valid PNG image",
			size:        2048,
			contentType: "image/png",
			wantError:   nil,
		},
		{
			name:        "valid WEBP image",
			size:        4096,
			contentType: "image/webp",
			wantError:   nil,
		},
		{
			name:        "image size exactly at maximum limit",
			size:        maxImageSize,
			contentType: "image/jpeg",
			wantError:   nil,
		},

		// Invalid size cases
		{
			name:        "zero image size",
			size:        0,
			contentType: "image/jpeg",
			wantError:   domainerrors.ErrInvalidProductImageSize,
		},
		{
			name:        "negative image size",
			size:        -1,
			contentType: "image/jpeg",
			wantError:   domainerrors.ErrInvalidProductImageSize,
		},
		{
			name:        "image size exceeds maximum limit",
			size:        maxImageSize + 1,
			contentType: "image/jpeg",
			wantError:   domainerrors.ErrInvalidProductImageSize,
		},

		// Invalid content type cases
		{
			name:        "GIF image not allowed",
			size:        1024,
			contentType: "image/gif",
			wantError:   domainerrors.ErrInvalidProductImage,
		},
		{
			name:        "PDF file not allowed",
			size:        1024,
			contentType: "application/pdf",
			wantError:   domainerrors.ErrInvalidProductImage,
		},
		{
			name:        "empty content type",
			size:        1024,
			contentType: "",
			wantError:   domainerrors.ErrInvalidProductImage,
		},

		// Multiple invalid conditions
		{
			name:        "invalid size and content type",
			size:        0,
			contentType: "application/pdf",
			wantError:   domainerrors.ErrInvalidProductImageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImage(tt.size, tt.contentType)

			if !errors.Is(err, tt.wantError) {
				t.Errorf(
					"validateImage() error = %v; want %v",
					err,
					tt.wantError,
				)
			}
		})
	}
}
