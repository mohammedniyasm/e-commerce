package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID int64) (*models.Cart, error)
	GetCartItems(ctx context.Context, cartID int64) ([]models.CartItem, error)
	CreateCart(ctx context.Context, cart *models.Cart) error
	GetCartItemByVariantID(ctx context.Context, cartID int64, variantID int64) (*models.CartItem, error)
	CreateCartItem(ctx context.Context, item *models.CartItem) error
	UpdateCartItemQuantity(ctx context.Context, cartID int64, itemID int64, quantity int) error
	RemoveCartItem(ctx context.Context, cartID int64, itemID int64) error
	ClearCart(ctx context.Context, cartID int64) error
	WithTransaction(ctx context.Context, fn func(txRepo CartRepository) error) error
}
