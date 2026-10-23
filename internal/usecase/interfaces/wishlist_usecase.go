package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type WishlistUseCase interface {
	GetWishlist(ctx context.Context, userID int64) (*models.Wishlist, error)
	AddProduct(ctx context.Context, userID int64, productID int64) (*models.Wishlist, error)
	RemoveItem(ctx context.Context, userID int64, itemID int64) (*models.Wishlist, error)
	ClearWishlist(ctx context.Context, userID int64) (*models.Wishlist, error)
	MergeGuestWishlist(ctx context.Context, userID int64, productIDs []int64) (*models.Wishlist, error)
}
