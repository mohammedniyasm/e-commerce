package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type WishlistRepository interface {
	GetWishlistByUserID(ctx context.Context, userID int64) (*models.Wishlist, error)
	CreateWishlist(ctx context.Context, wishlist *models.Wishlist) error
	GetWishlistItems(ctx context.Context, wishlistID int64) ([]models.WishlistItem, error)
	GetWishlistItemByProductID(ctx context.Context, wishlistID int64, productID int64) (*models.WishlistItem, error)
	CreateWishlistItem(ctx context.Context, item *models.WishlistItem) error
	RemoveWishlistItem(ctx context.Context, wishlistID int64, itemID int64) error
	ClearWishlist(ctx context.Context, wishlistID int64) error
	WithTransaction(ctx context.Context, fn func(txRepo WishlistRepository) error) error
}
