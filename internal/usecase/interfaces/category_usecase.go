package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type CategoryUseCase interface {
	CreateCategory(ctx context.Context,name string,description string) (*models.Category, error)
	GetCategory(ctx context.Context,id int64) (*models.Category, error)
	UpdateCategory(ctx context.Context,id int64,name string,description string) (*models.Category, error)
	DeleteCategory(ctx context.Context,id int64) error
	ListCategories(ctx context.Context,search string,page int,limit int) ([]models.Category, int64, error)
	RestoreCategory(ctx context.Context,id int64) error
	ListDeletedCategories(ctx context.Context,search string,page int,limit int) ([]models.Category, int64, error)
	ToggleCategoryActive(ctx context.Context,id int64) error
}

