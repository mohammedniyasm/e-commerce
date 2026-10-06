package interfaces

import (
	"context"
	"io"

	"ecommerce/internal/domain/models"
)

type BrandUseCase interface {
	CreateBrand(ctx context.Context, name string, description string) (*models.Brand, error)
	GetBrand(ctx context.Context, id int64) (*models.Brand, error)
	UpdateBrand(ctx context.Context, id int64, name string, description string) (*models.Brand, error)
	DeleteBrand(ctx context.Context, id int64) error
	ListBrands(ctx context.Context, search string, page int, limit int) ([]models.Brand, int64, error)
	RestoreBrand(ctx context.Context, id int64) error
	ListDeletedBrands(ctx context.Context, search string, page int, limit int) ([]models.Brand, int64, error)
	ToggleBrandActive(ctx context.Context, id int64) error
	UploadBrandLogo(ctx context.Context, id int64, file io.Reader, size int64, contentType string) (*models.Brand, error)
	DeleteBrandLogo(ctx context.Context, id int64) error
}
