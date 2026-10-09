package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, name string, shortDescription *string, description *string, categoryID int64, brandID int64) (*models.Product, error)
	GetProduct(ctx context.Context, id uint) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	UpdateProduct(ctx context.Context, id uint, name string, shortDescription *string, description *string, categoryID int64, brandID int64) (*models.Product, error)
	DeleteProduct(ctx context.Context, id uint) error
	ListProducts(ctx context.Context, search string, page int, limit int, isActive *bool, isListed *bool, categoryID *int64, brandID *int64) ([]models.Product, int64, error)
	ListDeletedProducts(ctx context.Context, search string, page int, limit int) ([]models.Product, int64, error)
	RestoreProduct(ctx context.Context, id uint) error
	ToggleProductActive(ctx context.Context, id uint) error
	ToggleProductListed(ctx context.Context, id uint) error
	ListStoreProducts(
		ctx context.Context,
		search string,
		categoryID *int64,
		brandID *int64,
		minPrice *float64,
		maxPrice *float64,
		sort string,
		page int,
		limit int,
	) ([]models.Product, int64, error)

	GetStoreProductByID(ctx context.Context, id int64) (*models.Product, error)
	GetStoreProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetRelatedProducts(ctx context.Context,productID int64,limit int) ([]models.Product, error)
}
