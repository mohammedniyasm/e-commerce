package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error

	GetByID(ctx context.Context, id uint) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetDeletedByID(ctx context.Context, id uint) (*models.Product, error)

	ToggleActive(ctx context.Context, id uint) error
	ToggleListed(ctx context.Context, id uint) error

	Update(ctx context.Context, product *models.Product) error
	SoftDelete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error

	List(ctx context.Context, search string, limit int, page int, isActive *bool, isListed *bool, categoryID *int64, brandID *int64) ([]models.Product, int64, error)
	ListDeleted(ctx context.Context, search string, limit int, page int) ([]models.Product, int64, error)

	CountActiveByCategoryID(ctx context.Context,categoryID int64) (int64, error)
	CountActiveByBrandID(ctx context.Context,brandID int64) (int64, error)
}
