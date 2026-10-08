package postgres

import (
	"context"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"

	"gorm.io/gorm"
)

type ProductVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) *ProductVariantRepository {
	return &ProductVariantRepository{
		db: db,
	}
}

func (r *ProductVariantRepository) Create(ctx context.Context,variant *models.ProductVariant) error {
	return r.db.WithContext(ctx).
		Create(variant).
		Error
}
func (r *ProductVariantRepository) GetByID(ctx context.Context,id int64) (*models.ProductVariant, error) {

	var variant models.ProductVariant

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&variant).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}

		return nil, err
	}

	return &variant, nil
}
func (r *ProductVariantRepository) GetBySKU(ctx context.Context,sku string) (*models.ProductVariant, error) {

	var variant models.ProductVariant

	err := r.db.WithContext(ctx).
		Where("sku = ?", sku).
		First(&variant).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}

		return nil, err
	}

	return &variant, nil
}
func (r *ProductVariantRepository) GetByProductAndAttributes(ctx context.Context,productID int64,size *string,color *string) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	query := r.db.WithContext(ctx).
		Where("product_id = ?", productID)
	if size == nil {
		query = query.Where("size IS NULL")
	} else {
		query = query.Where("size = ?", *size)
	}
	if color == nil {
		query = query.Where("color IS NULL")
	} else {
		query = query.Where("color = ?", *color)
	}
	err := query.First(&variant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &variant, nil
}
func (r *ProductVariantRepository) Update(ctx context.Context,variant *models.ProductVariant) error {

	return r.db.WithContext(ctx).
		Model(&models.ProductVariant{}).
		Where("id = ?", variant.ID).
		Updates(map[string]interface{}{
			"size":          variant.Size,
			"color":         variant.Color,
			"sku":           variant.SKU,
			"mrp":           variant.MRP,
			"selling_price": variant.SellingPrice,
			"stock":         variant.Stock,
			"is_active":     variant.IsActive,
		}).
		Error
}
func (r *ProductVariantRepository) Delete(ctx context.Context,id int64) error {
	return r.db.WithContext(ctx).
		Delete(&models.ProductVariant{}, id).
		Error
}
func (r *ProductVariantRepository) ListByProductID(ctx context.Context,productID int64) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("id ASC").
		Find(&variants).
		Error
	if err != nil {
		return nil, err
	}
	return variants, nil
}
