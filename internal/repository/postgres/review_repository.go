package postgres

import (
	"context"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"gorm.io/gorm"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{
		db: db,
	}
}
func (r *ReviewRepository) Create(ctx context.Context, review *models.Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}
func (r *ReviewRepository) GetByOrderItemID(ctx context.Context, orderItemID uint) (*models.Review, error) {
	var review models.Review
	err := r.db.WithContext(ctx).
		Where("order_item_id = ?", orderItemID).
		First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}

	return &review, nil
}
func (r *ReviewRepository) Update(ctx context.Context, reviewID uint, userID uint, rating int16, comment *string) error {
	result := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("id = ? AND user_id = ?", reviewID, userID).
		Updates(map[string]interface{}{
			"rating":  rating,
			"comment": comment,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}

	return nil
}
func (r *ReviewRepository) Delete(ctx context.Context, reviewID uint, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", reviewID, userID).
		Delete(&models.Review{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}

	return nil
}
func (r *ReviewRepository) GetByID(ctx context.Context, reviewID uint, userID uint) (*models.Review, error) {
	var review models.Review
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", reviewID, userID).
		First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &review, nil
}
func (r *ReviewRepository) ListByUserID(ctx context.Context, userID uint, page int, limit int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, err
	}
	return reviews, total, nil
}
func (r *ReviewRepository) ListByProductID(ctx context.Context, productID uint, page int, limit int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("product_id = ? AND is_visible = ?", productID, true)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, err
	}
	return reviews, total, nil
}
func (r *ReviewRepository) ListAll(ctx context.Context, page int, limit int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.Review{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Preload("Product").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, err
	}
	return reviews, total, nil
}
func (r *ReviewRepository) SetVisibility(ctx context.Context, reviewID uint, isVisible bool) error {
	result := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("id = ?", reviewID).
		Update("is_visible", isVisible)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}

var _ interfaces.ReviewRepository = (*ReviewRepository)(nil)
