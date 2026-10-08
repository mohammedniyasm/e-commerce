package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"errors"

	"gorm.io/gorm"
)

type BrandRepository struct {
	db *gorm.DB
}

func NewBrandRepository(db *gorm.DB) *BrandRepository {
	return &BrandRepository{
		db: db,
	}
}

func (r *BrandRepository) Create(ctx context.Context, brand *models.Brand) error {
	err := r.db.WithContext(ctx).Create(&brand).Error
	return err
}
func (r *BrandRepository) GetByID(ctx context.Context, id int64) (*models.Brand, error) {
	var brand *models.Brand
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&brand).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return brand, nil
}
func (r *BrandRepository) GetByName(ctx context.Context, name string) (*models.Brand, error) {
	var brand *models.Brand
	err := r.db.WithContext(ctx).Where("name ILIKE ? AND deleted_at IS NULL", "%"+name+"%").First(&brand).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return brand, nil
}
func (r *BrandRepository) GetByExactName(ctx context.Context, name string) (*models.Brand, error) {
	var brand models.Brand
	err := r.db.WithContext(ctx).
		Where(
			"LOWER(name) = LOWER(?) AND deleted_at IS NULL",
			name,
		).
		First(&brand).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}

	return &brand, nil
}
func (r *BrandRepository) Update(ctx context.Context, brand *models.Brand) error {
	return r.db.WithContext(ctx).
		Model(&models.Brand{}).
		Where("id = ? AND deleted_at IS NULL", brand.ID).
		Updates(map[string]interface{}{
			"name":        brand.Name,
			"description": brand.Description,
		}).Error
}
func (r *BrandRepository) UpdateLogo(ctx context.Context, id int64, logo *string) error {
	return r.db.WithContext(ctx).
		Model(&models.Brand{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("logo", logo).Error
}
func (r *BrandRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&models.Brand{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}
func (r *BrandRepository) List(ctx context.Context, search string, page int, limit int) ([]models.Brand, int64, error) {
	var count int64
	var brands []models.Brand
	query := r.db.WithContext(ctx).Model(&models.Brand{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.Order("created_at ASC").Offset(offset).Limit(limit).Find(&brands).Error
	if err != nil {
		return nil, 0, err
	}
	return brands, count, nil
}
func (r *BrandRepository) Restore(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&models.Brand{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Updates(map[string]interface{}{
			"is_active":  true,
			"deleted_at": nil,
		}).Error
}
func (r *BrandRepository) GetDeletedByID(ctx context.Context, id int64) (*models.Brand, error) {
	var brand models.Brand
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		First(&brand).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &brand, nil
}
func (r *BrandRepository) ListDeleted(ctx context.Context, search string, page int, limit int) ([]models.Brand, int64, error) {
	var brands []models.Brand
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Brand{}).Where("deleted_at IS NOT NULL")
	if search != "" {
		query = query.Where(
			"name ILIKE ?",
			"%"+search+"%",
		)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.Order("deleted_at DESC").Offset(offset).Limit(limit).Find(&brands).Error
	if err != nil {
		return nil, 0, err
	}
	return brands, total, nil
}
func (r *BrandRepository) ToggleActive(ctx context.Context, id int64) error {
	err := r.db.WithContext(ctx).Model(&models.Brand{}).Where("id = ? AND deleted_at IS NULL", id).Update("is_active", gorm.Expr("NOT is_active")).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainerrors.ErrRecordNotFound
		}
		return err
	}
	return err
}
