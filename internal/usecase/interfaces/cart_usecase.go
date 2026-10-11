package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type CartUseCase interface {
	GetCart(
		ctx context.Context,
		userID int64,
	) (*models.Cart, error)

	AddItem(
		ctx context.Context,
		userID int64,
		variantID int64,
		quantity int,
	) (*models.Cart, error)

	UpdateQuantity(
		ctx context.Context,
		userID int64,
		itemID int64,
		quantity int,
	) (*models.Cart, error)

	RemoveItem(
		ctx context.Context,
		userID int64,
		itemID int64,
	) (*models.Cart, error)

	ClearCart(
		ctx context.Context,
		userID int64,
	) (*models.Cart, error)
}
