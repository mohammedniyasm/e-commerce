package postgres

import (
	"context"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CouponUsageRepository struct {
	db *gorm.DB
}

func NewCouponUsageRepository(db *gorm.DB) *CouponUsageRepository {
	return &CouponUsageRepository{
		db: db,
	}
}
func (r *CouponUsageRepository) Create(ctx context.Context, usage *models.CouponUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}
func (r *CouponUsageRepository) CountByCouponID(ctx context.Context, couponID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CouponUsage{}).
		Where("coupon_id = ? AND status = ?", couponID, true).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (r *CouponUsageRepository) CountByCouponAndUser(ctx context.Context, couponID uuid.UUID, userID uint) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.CouponUsage{}).
		Where("coupon_id = ? AND user_id = ? AND status = ?", couponID, userID, true).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}
func (r *CouponUsageRepository) ListByCouponID(ctx context.Context, couponID uuid.UUID, limit int, page int) ([]models.CouponUsage, int64, error) {
	var usages []models.CouponUsage
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.CouponUsage{}).
		Where("coupon_id = ? AND status = ?", couponID, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	err := query.
		Preload("Coupon").
		Preload("Order").
		Order("used_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&usages).Error

	if err != nil {
		return nil, 0, err
	}

	return usages, total, nil
}
func (r *CouponUsageRepository) ListByUserID(ctx context.Context, userID uint, limit int, page int) ([]models.CouponUsage, int64, error) {
	var usages []models.CouponUsage
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.CouponUsage{}).
		Where("user_id = ? AND status = ?", userID, true)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.
		Preload("Coupon").
		Preload("Order").
		Order("used_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&usages).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, domainerrors.ErrRecordNotFound
		}
		return nil, 0, err
	}

	return usages, total, nil
}
func (r *CouponUsageRepository) CancelByOrderID(ctx context.Context, orderID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.CouponUsage{}).
		Where("order_id = ? AND status = ?", orderID, true).
		Update("status", false).Error
}

var _ interfaces.CouponUsageRepository = (*CouponUsageRepository)(nil)
