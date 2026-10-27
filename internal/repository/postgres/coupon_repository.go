package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{
		db: db,
	}
}
func (r *CouponRepository) Create(ctx context.Context, coupon *models.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}
func (r *CouponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Coupon{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *CouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &coupon, nil
}
func (r *CouponRepository) GetByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &coupon, nil
}
func (r *CouponRepository) Update(ctx context.Context, coupon *models.Coupon) error {
	result := r.db.WithContext(ctx).
		Model(&models.Coupon{}).
		Where("id = ?", coupon.ID).
		Select(
			"Code",
			"Description",
			"DiscountType",
			"DiscountValue",
			"MinimumOrderAmount",
			"MaximumDiscount",
			"UsageLimit",
			"UsageLimitPerUser",
			"StartAt",
			"ExpiresAt",
			"IsActive",
			"UpdatedAt",
		).Updates(coupon)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *CouponRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	result := r.db.WithContext(ctx).
		Model(&models.Coupon{}).
		Where("id = ?", id).
		Update("is_active", active)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *CouponRepository) List(ctx context.Context, search string, activeOnly *bool, limit int, page int, sort *string) ([]models.Coupon, int64, error) {
	var coupons []models.Coupon
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Coupon{})
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("code ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}
	if activeOnly != nil {
		query = query.Where("is_active = ?", *activeOnly)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if sort != nil {
		switch *sort {
		case "new":
			query = query.Order("created_at DESC")
		case "old":
			query = query.Order("created_at ASC")
		default:
			query = query.Order("created_at DESC")
		}
	}
	offset := (page - 1) * limit
	if err := query.Limit(limit).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}
func (r *CouponRepository) GetExpiredCoupons(ctx context.Context, now time.Time) ([]models.Coupon, error) {
	var coupons []models.Coupon
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("expires_at <= ?", now).
		Find(&coupons).Error
	if err != nil {
		return nil, err
	}
	return coupons, nil
}
func (r *CouponRepository) GetExhaustedCoupons(ctx context.Context) ([]models.Coupon, error) {
	var coupons []models.Coupon
	err := r.db.WithContext(ctx).
		Model(&models.Coupon{}).
		Where("is_active = ?", true).
		Where("usage_limit IS NOT NULL").
		Where(`
			usage_limit <= (
				SELECT COUNT(*)
				FROM coupon_usage
				WHERE coupon_usage.coupon_id = coupons.id
			)
		`).
		Find(&coupons).Error
	if err != nil {
		return nil, err
	}
	return coupons, nil
}
func (r *CouponRepository) GetAvailableCoupons(ctx context.Context, now time.Time) ([]models.Coupon, error) {
	var coupons []models.Coupon
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("start_at <= ?", now).
		Where("expires_at > ?", now).
		Find(&coupons).Error
	if err != nil {
		return nil, err
	}
	return coupons, nil
}
func (r *CouponRepository) GetByCodeForUpdate(ctx context.Context, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("code = ?", code).
		First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &coupon, nil
}
