package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type CouponUseCase interface {
	// Admin: Coupon management
	CreateCoupon(ctx context.Context, coupon *models.Coupon) (*models.Coupon, error)
	GetCoupon(ctx context.Context, couponID uuid.UUID) (*models.Coupon, error)
	UpdateCoupon(ctx context.Context, coupon *models.Coupon) (*models.Coupon, error)
	DeleteCoupon(ctx context.Context, couponID uuid.UUID) error
	ToggleCouponActive(ctx context.Context, couponID uuid.UUID) error
	ListCoupons(ctx context.Context, search string, activeOnly *bool, limit int, page int, sort *string) ([]models.Coupon, int64, error)
	// Customer: Coupon validation
	ValidateCoupon(ctx context.Context, code string, userID uint, orderAmount float64) (*models.Coupon, float64, error)
	// Coupon redemption
	RedeemCoupon(ctx context.Context, couponID uuid.UUID, userID uint, orderID uuid.UUID) error
	// Coupon usage history
	ListCouponUsagesByCouponID(ctx context.Context, couponID uuid.UUID, limit int, page int) ([]models.CouponUsage, int64, error)
	ListCouponUsagesByUserID(ctx context.Context, userID uint, limit int, page int) ([]models.CouponUsage, int64, error)
}
