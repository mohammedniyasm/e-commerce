package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"errors"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ? AND deleted_at is null", product.ID).Updates(map[string]interface{}{
		"name":              product.Name,
		"slug":              product.Slug,
		"short_description": product.ShortDescription,
		"description":       product.Description,
		"category_id":       product.CategoryID,
		"brand_id":          product.BrandID,
	}).Error
	return err
}
func (r *ProductRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}
func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}
func (r *ProductRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{
		"is_active":  false,
		"is_listed":  false,
		"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error
}
func (r *ProductRepository) List(ctx context.Context, search string, page int, limit int, isActive *bool, isListed *bool, categoryID *int64, brandID *int64) ([]models.Product, int64, error) {

	var products []models.Product
	var total int64

	countQuery := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("deleted_at IS NULL")

	if search != "" {
		searchPattern := "%" + search + "%"

		countQuery = countQuery.Where(
			"name ILIKE ? OR slug ILIKE ?",
			searchPattern,
			searchPattern,
		)
	}

	if isActive != nil {
		countQuery = countQuery.Where(
			"is_active = ?",
			*isActive,
		)
	}

	if isListed != nil {
		countQuery = countQuery.Where(
			"is_listed = ?",
			*isListed,
		)
	}

	if categoryID != nil {
		countQuery = countQuery.Where(
			"category_id = ?",
			*categoryID,
		)
	}

	if brandID != nil {
		countQuery = countQuery.Where(
			"brand_id = ?",
			*brandID,
		)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("deleted_at IS NULL")

	if search != "" {
		searchPattern := "%" + search + "%"

		dataQuery = dataQuery.Where(
			"name ILIKE ? OR slug ILIKE ?",
			searchPattern,
			searchPattern,
		)
	}

	if isActive != nil {
		dataQuery = dataQuery.Where(
			"is_active = ?",
			*isActive,
		)
	}

	if isListed != nil {
		dataQuery = dataQuery.Where(
			"is_listed = ?",
			*isListed,
		)
	}

	if categoryID != nil {
		dataQuery = dataQuery.Where(
			"category_id = ?",
			*categoryID,
		)
	}

	if brandID != nil {
		dataQuery = dataQuery.Where(
			"brand_id = ?",
			*brandID,
		)
	}

	offset := (page - 1) * limit

	err := dataQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&products).
		Error

	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
func (r *ProductRepository) ListDeleted(ctx context.Context, search string, page int, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("deleted_at IS NOT NULL")
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"name ILIKE ? OR slug ILIKE ?",
			searchPattern,
			searchPattern,
		)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.
		Order("deleted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&products).
		Error
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}
func (r *ProductRepository) GetDeletedByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Where(
			"id = ? AND deleted_at IS NOT NULL",
			id,
		).
		First(&product).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}
func (r *ProductRepository) Restore(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"id = ? AND deleted_at IS NOT NULL",
			id,
		).
		Updates(map[string]interface{}{
			"is_active":  true,
			"is_listed":  true,
			"deleted_at": nil,
		}).
		Error
}
func (r *ProductRepository) ToggleActive(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"id = ? AND deleted_at IS NULL",
			id,
		).
		Update(
			"is_active",
			gorm.Expr("NOT is_active"),
		).
		Error
}
func (r *ProductRepository) ToggleListed(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"id = ? AND deleted_at IS NULL",
			id,
		).
		Update(
			"is_listed",
			gorm.Expr("NOT is_listed"),
		).
		Error
}
func (r *ProductRepository) CountActiveByCategoryID(ctx context.Context, categoryID int64) (int64, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"category_id = ? AND is_active = true AND deleted_at IS NULL",
			categoryID,
		).
		Count(&count).
		Error

	return count, err
}
func (r *ProductRepository) CountActiveByBrandID(ctx context.Context,brandID int64) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"brand_id = ? AND is_active = true AND deleted_at IS NULL",
			brandID,
		).
		Count(&count).
		Error

	return count, err
}
