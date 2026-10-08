package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"errors"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	err := r.db.WithContext(ctx).Create(&category).Error
	return err
}
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	var category *models.Category
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return category, nil
}
func (r *CategoryRepository) GetByName(ctx context.Context, name string) (*models.Category, error) {
	var category *models.Category
	err := r.db.WithContext(ctx).Where("name ILIKE ? AND deleted_at IS NULL", "%"+name+"%").First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return category, nil
}
func (r *CategoryRepository) GetByExactName(ctx context.Context,name string) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Where(
			"LOWER(name) = LOWER(?) AND deleted_at IS NULL",
			name,
		).
		First(&category).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}

	return &category, nil
}
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where("id = ? AND deleted_at IS NULL", category.ID).
		Updates(map[string]interface{}{
			"name":        category.Name,
			"description": category.Description,
		}).Error
}
func (r *CategoryRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}
func (r *CategoryRepository) List(ctx context.Context, search string, page int, limit int) ([]models.Category, int64, error) {
	var count int64
	var categories []models.Category
	query := r.db.WithContext(ctx).Model(&models.Category{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.Order("created_at ASC").Offset(offset).Limit(limit).Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}
	return categories, count, nil
}
func (r *CategoryRepository) Restore(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Updates(map[string]interface{}{
			"is_active":  true,
			"deleted_at": nil,
		}).Error
}
func (r *CategoryRepository) GetDeletedByID(ctx context.Context, id int64) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &category, nil
}
func (r *CategoryRepository) ListDeleted(ctx context.Context, search string, page int, limit int) ([]models.Category, int64, error) {
	var categories []models.Category
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Category{}).Where("deleted_at IS NOT NULL")
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
	err := query.Order("deleted_at DESC").Offset(offset).Limit(limit).Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}
	return categories, total, nil
}
func (r *CategoryRepository) ToggleActive(ctx context.Context, id int64) error {
	err := r.db.WithContext(ctx).Model(&models.Category{}).Where("id = ? AND deleted_at IS NULL", id).Update("is_active", gorm.Expr("NOT is_active")).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainerrors.ErrRecordNotFound
		}
		return err
	}
	return err
}
