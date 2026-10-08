package interfaces

import (
	"context"
	"io"

	"ecommerce/internal/domain/models"
)

type ProductImageUpload struct {
	File        io.Reader
	Size        int64
	ContentType string
}

type ProductImageUseCase interface {
	UploadProductImages(ctx context.Context, productID int64, variantID *int64, files []ProductImageUpload) ([]models.ProductImage, error)
	GetProductImage(ctx context.Context, id int64) (*models.ProductImage, error)
	ListProductImages(ctx context.Context, productID int64) ([]models.ProductImage, error)
	ListVariantImages(ctx context.Context, variantID int64) ([]models.ProductImage, error)
	UpdateProductImage(ctx context.Context, id int64, variantID *int64, displayOrder *int, isPrimary *bool) (*models.ProductImage, error)
	DeleteProductImage(ctx context.Context, id int64) error
}
