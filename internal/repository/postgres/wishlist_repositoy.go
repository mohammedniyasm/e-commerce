package postgres

import (
	"context"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"gorm.io/gorm"
)

type wishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) *wishlistRepository {
	return &wishlistRepository{
		db: db,
	}
}
func (r *wishlistRepository) GetWishlistByUserID(ctx context.Context, userID int64) (*models.Wishlist, error) {
	var wishlist models.Wishlist
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&wishlist).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &wishlist, nil
}
func (r *wishlistRepository) CreateWishlist(ctx context.Context, wishlist *models.Wishlist) error {
	return r.db.WithContext(ctx).Create(wishlist).Error
}
func (r *wishlistRepository) GetWishlistItems(ctx context.Context, wishlistID int64) ([]models.WishlistItem, error) {
	items := make([]models.WishlistItem, 0)
	err := r.db.WithContext(ctx).
		Where("wishlist_id = ?", wishlistID).
		Preload("Product").
		Preload("Product.Category").
		Preload("Product.Brand").
		Preload(
			"Product.Images",
			"is_primary = ?",
			true,
		).
		Order("wishlist_items.created_at DESC").
		Order("wishlist_items.id DESC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
func (r *wishlistRepository) GetWishlistItemByProductID(ctx context.Context, wishlistID int64, productID int64) (*models.WishlistItem, error) {
	var wishlistItem models.WishlistItem
	err := r.db.WithContext(ctx).Where("wishlist_id = ? AND product_id = ?", wishlistID, productID).First(&wishlistItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &wishlistItem, nil
}
func (r *wishlistRepository) CreateWishlistItem(ctx context.Context, item *models.WishlistItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *wishlistRepository) RemoveWishlistItem(ctx context.Context, wishlistID int64, itemID int64) error {
	err := r.db.WithContext(ctx).Where("wishlist_id = ?", wishlistID).Where("id = ?", itemID).Delete(&models.WishlistItem{})
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *wishlistRepository) ClearWishlist(ctx context.Context, wishlistID int64) error {
	return r.db.WithContext(ctx).Where("wishlist_id = ?", wishlistID).Delete(&models.WishlistItem{}).Error
}
func (r *wishlistRepository) WithTransaction(ctx context.Context, fn func(txRepo interfaces.WishlistRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &wishlistRepository{
			db: tx,
		}
		return fn(txRepo)
	})
}
