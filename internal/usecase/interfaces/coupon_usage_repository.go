package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type CouponUsageRepository interface {
	Create(ctx context.Context, usage *models.CouponUsage) error
	CountByCouponID(ctx context.Context, couponID uuid.UUID) (int64, error)
	CountByCouponAndUser(ctx context.Context, couponID uuid.UUID, userID uint) (int64, error)
	ListByCouponID(ctx context.Context, couponID uuid.UUID, limit int, page int) ([]models.CouponUsage, int64, error)
	ListByUserID(ctx context.Context, userID uint, limit int, page int) ([]models.CouponUsage, int64, error)
	CancelByOrderID(ctx context.Context, orderID uuid.UUID) error
}
