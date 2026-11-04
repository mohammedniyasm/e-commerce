package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *models.Review) error
	GetByOrderItemID(ctx context.Context, orderItemID uint) (*models.Review, error)
	Update(ctx context.Context, reviewID uint, userID uint, rating int16, comment *string) error
	Delete(ctx context.Context, reviewID uint, userID uint) error
	GetByID(ctx context.Context, reviewID uint, userID uint) (*models.Review, error)
	ListByUserID(ctx context.Context, userID uint, page int, limit int) ([]models.Review, int64, error)
	ListByProductID(ctx context.Context, productID uint, page int, limit int) ([]models.Review, int64, error)
	ListAll(ctx context.Context, page int, limit int) ([]models.Review, int64, error)
	SetVisibility(ctx context.Context, reviewID uint, isVisible bool) error
}
