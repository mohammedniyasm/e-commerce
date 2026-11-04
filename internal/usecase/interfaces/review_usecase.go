package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type ReviewUseCase interface {
	// Customer operations
	CreateReview(ctx context.Context, userID uint, orderItemID uint, rating int16, comment *string) (*models.Review, error)
	GetMyReview(ctx context.Context, userID uint, reviewID uint) (*models.Review, error)
	ListMyReviews(ctx context.Context, userID uint, page int, limit int) ([]models.Review, int64, error)
	UpdateReview(ctx context.Context, userID uint, reviewID uint, rating int16, comment *string) error
	DeleteReview(ctx context.Context, userID uint, reviewID uint) error
	// Public operation
	ListProductReviews(ctx context.Context, productID uint, page int, limit int) ([]models.Review, int64, error)
	// Admin operations
	ListAllReviews(ctx context.Context, page int, limit int) ([]models.Review, int64, error)
	SetReviewVisibility(ctx context.Context, reviewID uint, isVisible bool) error
}
