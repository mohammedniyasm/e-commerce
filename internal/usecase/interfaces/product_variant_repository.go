package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type ProductVariantRepository interface {
	Create(ctx context.Context, variant *models.ProductVariant) error
	GetByID(ctx context.Context, id int64) (*models.ProductVariant, error)
	GetBySKU(ctx context.Context, sku string) (*models.ProductVariant, error)
	GetByProductAndAttributes(ctx context.Context, productID int64, size *string, color *string) (*models.ProductVariant, error)
	Update(ctx context.Context, variant *models.ProductVariant) error
	Delete(ctx context.Context, id int64) error
	ListByProductID(ctx context.Context, productID int64) ([]models.ProductVariant, error)
	GetByIDs(ctx context.Context, ids []int64) ([]models.ProductVariant, error)
	DecreaseStock(ctx context.Context, variantID int64, quantity int) error
	IncreaseStock(ctx context.Context, variantID int64, quantity int) error
}
