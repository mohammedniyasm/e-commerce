package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type CouponRepository interface {
	Create(ctx context.Context, coupon *models.Coupon) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Coupon, error)
	GetByCode(ctx context.Context, code string) (*models.Coupon, error)
	Update(ctx context.Context, coupon *models.Coupon) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	List(ctx context.Context, search string, activeOnly *bool, limit int, page int, sort *string) ([]models.Coupon, int64, error)
	GetExpiredCoupons(ctx context.Context, now time.Time) ([]models.Coupon, error)
	GetExhaustedCoupons(ctx context.Context) ([]models.Coupon, error)
	GetAvailableCoupons(ctx context.Context, now time.Time) ([]models.Coupon, error)
	GetByCodeForUpdate(ctx context.Context, code string) (*models.Coupon, error)
}
