package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type ProductVariantUseCase interface {
	CreateVariant(ctx context.Context, productID int64, size *string, color *string, sku string, mrp float64, sellingPrice float64, stock int) (*models.ProductVariant, error)
	GetVariant(ctx context.Context, id int64) (*models.ProductVariant, error)
	UpdateVariant(ctx context.Context, id int64, size *string, color *string, sku string, mrp float64, sellingPrice float64, stock int, isActive *bool) (*models.ProductVariant, error)
	DeleteVariant(ctx context.Context, id int64) error
	ListVariantsByProductID(ctx context.Context, productID int64) ([]models.ProductVariant, error)
}
