package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type ProductImageRepository interface {
	Create(ctx context.Context, image *models.ProductImage) error

	GetByID(ctx context.Context, id int64) (*models.ProductImage, error)

	ListByProductID(ctx context.Context, productID int64) ([]models.ProductImage, error)
	ListByVariantID(ctx context.Context, variantID int64) ([]models.ProductImage, error)

	Update(ctx context.Context, image *models.ProductImage) error
	Delete(ctx context.Context, id int64) error

	DeleteByProductID(ctx context.Context, productID int64) error
	DeleteByVariantID(ctx context.Context, variantID int64) error

	CountByProductID(ctx context.Context, productID int64) (int64, error)
	CountByVariantID(ctx context.Context, variantID int64) (int64, error)
	ClearPrimaryByProductID(ctx context.Context, productID int64, excludeID int64) error
	UpdateDisplayOrder(ctx context.Context, id int64, newOrder int) error
}
