package couponuseCase

import (
	"context"
	"log/slog"
	"strings"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
)

type CouponUseCase struct {
	couponRepo      interfaces.CouponRepository
	orderRepo       interfaces.OrderRepository
	couponUsageRepo interfaces.CouponUsageRepository
	log             *slog.Logger
}

func NewCouponUseCase(
	couponRepo interfaces.CouponRepository,
	orderRepo interfaces.OrderRepository,
	couponUsageRepo interfaces.CouponUsageRepository,
	log *slog.Logger,
) *CouponUseCase {
	return &CouponUseCase{
		couponRepo:      couponRepo,
		orderRepo:       orderRepo,
		couponUsageRepo: couponUsageRepo,
		log:             log,
	}
}

func (u *CouponUseCase) CreateCoupon(ctx context.Context, coupon *models.Coupon) (*models.Coupon, error) {
	if coupon == nil {
		return nil, domainerrors.ErrInvalidCoupon
	}
	coupon.Code = strings.ToUpper(strings.TrimSpace(coupon.Code))
	if err := validateCoupon(coupon); err != nil {
		return nil, err
	}
	// Check whether the coupon code is already used.
	existing, err := u.couponRepo.GetByCode(ctx, coupon.Code)
	if err == nil && existing != nil {
		return nil, domainerrors.ErrCouponAlreadyExists
	}
	if err != nil && err != domainerrors.ErrRecordNotFound {
		return nil, err
	}
	if coupon.ID == uuid.Nil {
		coupon.ID = uuid.New()
	}
	now := time.Now()
	coupon.CreatedAt = now
	coupon.UpdatedAt = now
	if err := u.couponRepo.Create(ctx, coupon); err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to create coupon",
			"coupon_code", coupon.Code,
			"error", err,
		)
		return nil, err
	}
	u.log.InfoContext(
		ctx,
		"coupon created",
		"coupon_id", coupon.ID,
		"coupon_code", coupon.Code,
	)
	return coupon, nil
}
func (u *CouponUseCase) GetCoupon(ctx context.Context, couponID uuid.UUID) (*models.Coupon, error) {
	if couponID == uuid.Nil {
		return nil, domainerrors.ErrInvalidCouponID
	}
	return u.couponRepo.GetByID(ctx, couponID)
}
func (u *CouponUseCase) UpdateCoupon(ctx context.Context, coupon *models.Coupon) (*models.Coupon, error) {
	if coupon == nil || coupon.ID == uuid.Nil {
		return nil, domainerrors.ErrInvalidCouponID
	}
	coupon.Code = strings.ToUpper(strings.TrimSpace(coupon.Code))
	if err := validateCoupon(coupon); err != nil {
		return nil, err
	}
	existing, err := u.couponRepo.GetByID(ctx, coupon.ID)
	if err != nil {
		return nil, err
	}
	byCode, err := u.couponRepo.GetByCode(ctx, coupon.Code)
	if err == nil && byCode.ID != coupon.ID {
		return nil, domainerrors.ErrCouponAlreadyExists
	}
	if err != nil && err != domainerrors.ErrRecordNotFound {
		return nil, err
	}
	coupon.CreatedAt = existing.CreatedAt
	coupon.UpdatedAt = time.Now()
	if err := u.couponRepo.Update(ctx, coupon); err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to update coupon",
			"coupon_id", coupon.ID,
			"error", err,
		)
		return nil, err
	}
	u.log.InfoContext(
		ctx,
		"coupon updated",
		"coupon_id", coupon.ID,
	)
	return coupon, nil
}
func (u *CouponUseCase) DeleteCoupon(ctx context.Context, couponID uuid.UUID) error {
	if couponID == uuid.Nil {
		return domainerrors.ErrInvalidCouponID
	}
	if err := u.couponRepo.Delete(ctx, couponID); err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to delete coupon",
			"coupon_id", couponID,
			"error", err,
		)
		return err
	}
	u.log.InfoContext(
		ctx,
		"coupon deleted",
		"coupon_id", couponID,
	)
	return nil
}
func (u *CouponUseCase) ToggleCouponActive(ctx context.Context, couponID uuid.UUID) error {
	if couponID == uuid.Nil {
		return domainerrors.ErrInvalidCouponID
	}
	coupon, err := u.couponRepo.GetByID(ctx, couponID)
	if err != nil {
		return err
	}
	return u.couponRepo.SetActive(
		ctx,
		couponID,
		!coupon.IsActive,
	)
}
func (u *CouponUseCase) ListCoupons(ctx context.Context, search string, activeOnly *bool, limit int, page int, sort *string) ([]models.Coupon, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	coupons, total, err := u.couponRepo.List(
		ctx,
		strings.TrimSpace(search),
		activeOnly,
		limit,
		page,
		sort,
	)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to list coupons",
			"error", err,
		)
		return nil, 0, err
	}
	return coupons, total, nil
}
func (u *CouponUseCase) ValidateCoupon(ctx context.Context, code string, userID uint, orderAmount float64) (*models.Coupon, float64, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" || userID == 0 || orderAmount < 0 {
		return nil, 0, domainerrors.ErrInvalidInput
	}
	coupon, err := u.couponRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, 0, err
	}
	now := time.Now()
	if !coupon.IsActive {
		return nil, 0, domainerrors.ErrCouponNotActive
	}
	if now.Before(coupon.StartAt) {
		return nil, 0, domainerrors.ErrCouponNotStarted
	}
	if !now.Before(coupon.ExpiresAt) {
		return nil, 0, domainerrors.ErrCouponExpired
	}
	if orderAmount < coupon.MinimumOrderAmount {
		return nil, 0, domainerrors.ErrCouponMinimumOrderAmount
	}
	if coupon.UsageLimit != nil {
		used, err := u.couponUsageRepo.CountByCouponID(
			ctx,
			coupon.ID,
		)
		if err != nil {
			return nil, 0, err
		}
		if used >= int64(*coupon.UsageLimit) {
			return nil, 0, domainerrors.ErrCouponUsageLimitReached
		}
	}
	if coupon.UsageLimitPerUser != nil {
		used, err := u.couponUsageRepo.CountByCouponAndUser(
			ctx,
			coupon.ID,
			userID,
		)
		if err != nil {
			return nil, 0, err
		}

		if used >= int64(*coupon.UsageLimitPerUser) {
			return nil, 0, domainerrors.ErrCouponUsageLimitReached
		}
	}
	discount, err := calculateDiscount(coupon, orderAmount)
	if err != nil {
		return nil, 0, err
	}
	return coupon, discount, nil
}
func (u *CouponUseCase) RedeemCoupon(ctx context.Context, couponID uuid.UUID, userID uint, orderID uuid.UUID) error {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	coupon, err := u.couponRepo.GetByID(ctx, couponID)
	if err != nil {
		return err
	}
	_, discountAmount, err := u.ValidateCoupon(ctx, coupon.Code, userID, order.Subtotal)
	if err != nil {
		return err
	}
	now := time.Now()
	usage := &models.CouponUsage{
		CouponID:       couponID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmount: discountAmount,
		UsedAt:         now,
	}
	if err := u.couponUsageRepo.Create(ctx, usage); err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to redeem coupon",
			"coupon_id", couponID,
			"order_id", orderID,
			"error", err,
		)
		return err
	}
	u.log.InfoContext(
		ctx,
		"coupon redeemed",
		"coupon_id", couponID,
		"user_id", userID,
		"order_id", orderID,
	)
	return nil
}
func (u *CouponUseCase) ListCouponUsagesByCouponID(ctx context.Context, couponID uuid.UUID, limit int, page int) ([]models.CouponUsage, int64, error) {
	if couponID == uuid.Nil {
		return nil, 0, domainerrors.ErrInvalidCouponID
	}
	limit, page = normalizePagination(limit, page)
	return u.couponUsageRepo.ListByCouponID(
		ctx,
		couponID,
		limit,
		page,
	)
}
func (u *CouponUseCase) ListCouponUsagesByUserID(ctx context.Context, userID uint, limit int, page int) ([]models.CouponUsage, int64, error) {
	if userID == 0 {
		return nil, 0, domainerrors.ErrInvalidUserID
	}
	limit, page = normalizePagination(limit, page)
	return u.couponUsageRepo.ListByUserID(
		ctx,
		userID,
		limit,
		page,
	)
}
func validateCoupon(coupon *models.Coupon) error {
	if coupon.Code == "" {
		return domainerrors.ErrInvalidCouponCode
	}
	switch coupon.DiscountType {
	case "percentage":
		if coupon.DiscountValue <= 0 || coupon.DiscountValue > 100 {
			return domainerrors.ErrCouponDiscountInvalid
		}
	case "fixed":
		if coupon.DiscountValue <= 0 {
			return domainerrors.ErrCouponDiscountInvalid
		}
	default:
		return domainerrors.ErrCouponDiscountInvalid
	}
	if coupon.MinimumOrderAmount < 0 {
		return domainerrors.ErrInvalidCoupon
	}
	if coupon.MaximumDiscount != nil && *coupon.MaximumDiscount <= 0 {
		return domainerrors.ErrInvalidCoupon
	}
	if coupon.UsageLimit != nil && *coupon.UsageLimit <= 0 {
		return domainerrors.ErrInvalidCoupon
	}
	if coupon.UsageLimitPerUser != nil && *coupon.UsageLimitPerUser <= 0 {
		return domainerrors.ErrInvalidCoupon
	}
	if !coupon.ExpiresAt.After(coupon.StartAt) {
		return domainerrors.ErrInvalidCoupon
	}
	return nil
}
func calculateDiscount(coupon *models.Coupon, orderAmount float64) (float64, error) {
	var discount float64
	switch coupon.DiscountType {
	case "percentage":
		discount = orderAmount * coupon.DiscountValue / 100
	case "fixed":
		discount = coupon.DiscountValue
	default:
		return 0, domainerrors.ErrCouponDiscountInvalid
	}
	if coupon.MaximumDiscount != nil &&
		discount > *coupon.MaximumDiscount {
		discount = *coupon.MaximumDiscount
	}
	if discount > orderAmount {
		discount = orderAmount
	}
	return discount, nil
}
func normalizePagination(limit, page int) (int, int) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	return limit, page
}

var _ interfaces.CouponUseCase = (*CouponUseCase)(nil)
