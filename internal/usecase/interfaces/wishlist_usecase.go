package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type WishlistUseCase interface {
	CreateWishlist(ctx context.Context, userID uint, name string) error
	GetWishlists(ctx context.Context, userID uint) ([]models.Wishlist, error)
	GetWishlistByID(ctx context.Context, userID uint, wishlistID uint) (*models.Wishlist, error)
	UpdateWishlist(ctx context.Context, wishlistID uint, userID uint, name string) error
	DeleteWishlist(ctx context.Context, wishlistID uint, userID uint) error
	AddProductToWishlist(ctx context.Context, wishlistName string, userID uint, productID uint) (*models.Wishlist, error)
	RemoveWishlistItem(ctx context.Context, wishlistID uint, userID uint, itemID uint) (*models.Wishlist, error)
	ClearWishlist(ctx context.Context, wishlistID uint, userID uint) (*models.Wishlist, error)
	MergeWishlistItems(ctx context.Context, wishlistName string, userID uint, productIDs []uint) (*models.Wishlist, uint, uint, error)
}
