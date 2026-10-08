package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id int64) (*models.Category, error)
	GetByName(ctx context.Context, name string) (*models.Category, error)
	GetByExactName(ctx context.Context, name string) (*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	SoftDelete(ctx context.Context, id int64) error
	List(ctx context.Context, search string, page int, limit int) ([]models.Category, int64, error)
	Restore(ctx context.Context, id int64) error
	ToggleActive(ctx context.Context, id int64) error
	ListDeleted(ctx context.Context, search string, page int, limit int) ([]models.Category, int64, error)
	GetDeletedByID(ctx context.Context, id int64) (*models.Category, error)
}
