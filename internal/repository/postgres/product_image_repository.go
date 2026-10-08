package postgres

import (
	"context"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"

	"gorm.io/gorm"
)

type ProductImageRepository struct {
	db *gorm.DB
}

func NewProductImageRepository(db *gorm.DB) *ProductImageRepository {
	return &ProductImageRepository{
		db: db,
	}
}

func (r *ProductImageRepository) Create(ctx context.Context, image *models.ProductImage) error {

	return r.db.WithContext(ctx).
		Create(image).
		Error
}
func (r *ProductImageRepository) GetByID(ctx context.Context, id int64) (*models.ProductImage, error) {

	var image models.ProductImage

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&image).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}

		return nil, err
	}

	return &image, nil
}
func (r *ProductImageRepository) ListByProductID(ctx context.Context, productID int64) ([]models.ProductImage, error) {

	var images []models.ProductImage

	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("display_order ASC, id ASC").
		Find(&images).
		Error

	if err != nil {
		return nil, err
	}

	return images, nil
}
func (r *ProductImageRepository) ListByVariantID(ctx context.Context, variantID int64) ([]models.ProductImage, error) {

	var images []models.ProductImage

	err := r.db.WithContext(ctx).
		Where("variant_id = ?", variantID).
		Order("display_order ASC, id ASC").
		Find(&images).
		Error

	if err != nil {
		return nil, err
	}

	return images, nil
}
func (r *ProductImageRepository) Update(ctx context.Context, image *models.ProductImage) error {
	return r.db.WithContext(ctx).
		Model(&models.ProductImage{}).
		Where("id = ?", image.ID).
		Updates(map[string]interface{}{
			"image_url":     image.ImageURL,
			"variant_id":    image.VariantID,
			"display_order": image.DisplayOrder,
			"is_primary":    image.IsPrimary,
		}).
		Error
}
func (r *ProductImageRepository) Delete(ctx context.Context, id int64) error {

	result := r.db.WithContext(ctx).
		Delete(&models.ProductImage{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}

	return nil
}
func (r *ProductImageRepository) DeleteByProductID(ctx context.Context, productID int64) error {

	return r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Delete(&models.ProductImage{}).
		Error
}
func (r *ProductImageRepository) DeleteByVariantID(ctx context.Context, variantID int64) error {

	return r.db.WithContext(ctx).
		Where("variant_id = ?", variantID).
		Delete(&models.ProductImage{}).
		Error
}
func (r *ProductImageRepository) CountByProductID(ctx context.Context, productID int64) (int64, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.ProductImage{}).
		Where("product_id = ?", productID).
		Count(&count).
		Error

	return count, err
}
func (r *ProductImageRepository) CountByVariantID(ctx context.Context, variantID int64) (int64, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.ProductImage{}).
		Where("variant_id = ?", variantID).
		Count(&count).
		Error

	return count, err

}
func (r *ProductImageRepository) ClearPrimaryByProductID(ctx context.Context, productID int64, excludeID int64) error {

	return r.db.WithContext(ctx).
		Model(&models.ProductImage{}).
		Where(
			"product_id = ? AND id <> ?",
			productID,
			excludeID,
		).
		Update("is_primary", false).
		Error
}
func (r *ProductImageRepository) UpdateDisplayOrder(ctx context.Context, id int64, newOrder int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var image models.ProductImage
		if err := tx.
			Where("id = ?", id).
			First(&image).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerrors.ErrRecordNotFound
			}
			return err
		}
		oldOrder := image.DisplayOrder
		if oldOrder == newOrder {
			return nil
		}
		var targetImage models.ProductImage

		err := tx.Where("product_id = ? AND display_order = ?", image.ProductID, newOrder).First(&targetImage).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Model(&models.ProductImage{}).Where("id = ?", id).Update("display_order", newOrder).Error; err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&models.ProductImage{}).Where("id = ?", targetImage.ID).Update("display_order", oldOrder).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.ProductImage{}).Where("id = ?", id).Update("display_order", newOrder).Error; err != nil {
			return err
		}
		return nil
	})
}
