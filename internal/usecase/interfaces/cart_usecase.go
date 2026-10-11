package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type GuestCartItemInput struct {
	VariantID int64
	Quantity  int
}
type GuestCartItemResult struct {
	VariantID          int64
	RequestedQuantity  int
	AvailableStock     int
	CurrentPrice       float64
	IsAvailable        bool
	AvailabilityReason string
}
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
	ValidateGuestCart(ctx context.Context, items []GuestCartItemInput) ([]GuestCartItemResult, error)
	MergeGuestCart(ctx context.Context,userID int64,items []GuestCartItemInput) (*models.Cart, error)
}
