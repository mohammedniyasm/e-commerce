package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type BrandRepository interface {
	Create(ctx context.Context, category *models.Brand) error
	GetByID(ctx context.Context, id int64) (*models.Brand, error)
	GetByName(ctx context.Context, name string) (*models.Brand, error)
	GetByExactName(ctx context.Context, name string) (*models.Brand, error)
	Update(ctx context.Context, category *models.Brand) error
	SoftDelete(ctx context.Context, id int64) error
	UpdateLogo(ctx context.Context, id int64, logo *string) error
	List(ctx context.Context, search string, page int, limit int) ([]models.Brand, int64, error)
	Restore(ctx context.Context, id int64) error
	ToggleActive(ctx context.Context, id int64) error
	ListDeleted(ctx context.Context, search string, page int, limit int) ([]models.Brand, int64, error)
	GetDeletedByID(ctx context.Context, id int64) (*models.Brand, error)
}
