package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
)

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) interfaces.CartRepository {
	return &cartRepository{
		db: db,
	}
}

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID int64) (*models.Cart, error) {
	var cart models.Cart

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&cart).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}

		return nil, err
	}

	return &cart, nil
}
func (r *cartRepository) GetCartItems(ctx context.Context, cartID int64) ([]models.CartItem, error) {
	var items []models.CartItem

	err := r.db.WithContext(ctx).
		Where("cart_id = ?", cartID).
		Preload("Variant").
		Preload("Variant.Product").
		Preload("Variant.Product.Category").
		Preload("Variant.Product.Brand").
		Preload("Variant.Images").
		Preload(
			"Variant.Product.Images",
			"is_primary = ?",
			true,
		).
		Order("cart_items.created_at ASC").
		Order("cart_items.id ASC").
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}
func (r *cartRepository) CreateCart(ctx context.Context, cart *models.Cart) error {
	if err := r.db.WithContext(ctx).Create(cart).Error; err != nil {
		return err
	}

	return nil
}
func (r *cartRepository) GetCartItemByVariantID(ctx context.Context, cartID int64, variantID int64) (*models.CartItem, error) {
	var item models.CartItem

	err := r.db.WithContext(ctx).
		Where("cart_id = ?", cartID).
		Where("variant_id = ?", variantID).
		First(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}

		return nil, err
	}

	return &item, nil
}
func (r *cartRepository) CreateCartItem(ctx context.Context, item *models.CartItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return err
	}

	return nil
}
func (r *cartRepository) UpdateCartItemQuantity(ctx context.Context, cartID int64, itemID int64, quantity int) error {
	result := r.db.WithContext(ctx).
		Model(&models.CartItem{}).
		Where("id = ?", itemID).
		Where("cart_id = ?", cartID).
		Updates(map[string]interface{}{
			"quantity":   quantity,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}

	return nil
}
func (r *cartRepository) RemoveCartItem(ctx context.Context, cartID int64, itemID int64) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", itemID).
		Where("cart_id = ?", cartID).
		Delete(&models.CartItem{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}

	return nil
}
func (r *cartRepository) ClearCart(ctx context.Context, cartID int64) error {
	result := r.db.WithContext(ctx).
		Where("cart_id = ?", cartID).
		Delete(&models.CartItem{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}
func (r *cartRepository) WithTransaction(ctx context.Context, fn func(txRepo interfaces.CartRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &cartRepository{
			db: tx,
		}
		return fn(txRepo)
	})
}
