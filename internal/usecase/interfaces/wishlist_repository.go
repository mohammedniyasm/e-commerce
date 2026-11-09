package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type WishlistRepository interface {
	CreateWishlist(ctx context.Context, wishlist *models.Wishlist) error
	UpdateWishlist(ctx context.Context, wishlistID uint, userID uint, name string) error
	DeleteWishlist(ctx context.Context, wishlistID uint, userID uint) error
	GetWishlistByID(ctx context.Context, userID uint, wishlistID uint) (*models.Wishlist, error)
	GetWishlistByName(ctx context.Context, userID uint, wishlistName string) (*models.Wishlist, error)
	GetWishlistsByUserID(ctx context.Context, userID uint) ([]models.Wishlist, error)

	AddProductToWishlist(ctx context.Context, wishlistname string, userID uint, productID uint) error
	RemoveWishlistItem(ctx context.Context, wishlistID uint, userID uint, itemID uint) error
	ClearWishlist(ctx context.Context, wishlistID uint, userID uint) error

	MergeWishlistItems(ctx context.Context, wishlistName string, userID uint, productIDs []uint) (uint, uint, error)
}
