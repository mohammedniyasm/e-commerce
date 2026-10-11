package cart

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"fmt"
	"log/slog"
)

type CartUseCase struct {
	cartRepo    interfaces.CartRepository
	variantRepo interfaces.ProductVariantRepository
	log         *slog.Logger
}

func NewCartUseCase(
	cartRepo interfaces.CartRepository,
	variantRepo interfaces.ProductVariantRepository,
	log *slog.Logger,
) *CartUseCase {
	return &CartUseCase{
		cartRepo:    cartRepo,
		variantRepo: variantRepo,
		log:         log,
	}
}
func (u *CartUseCase) getOrCreateCart(ctx context.Context, userID int64) (*models.Cart, error) {
	cart, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	newCart := &models.Cart{
		UserID: uint(userID),
	}

	if err := u.cartRepo.CreateCart(ctx, newCart); err != nil {
		return nil, fmt.Errorf("create cart: %w", err)
	}

	return newCart, nil
}
func (u *CartUseCase) GetCart(ctx context.Context, userID int64) (*models.Cart, error) {
	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := u.cartRepo.GetCartItems(
		ctx,
		int64(cart.ID),
	)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	cart.Items = items

	return cart, nil
}
func validateVariant(variant *models.ProductVariant, quantity int) error {
	if variant == nil || variant.ID == 0 {
		return domainerrors.ErrRecordNotFound
	}
	if !variant.IsActive || variant.Stock <= 0 {
		return domainerrors.ErrProductNotAvailable
	}
	if variant.Product.ID == 0 ||
		!variant.Product.IsActive ||
		!variant.Product.IsListed ||
		variant.Product.DeletedAt != nil {
		return domainerrors.ErrProductNotAvailable
	}

	if quantity <= 0 {
		return domainerrors.ErrInvalidQuantity
	}

	if quantity > variant.Stock {
		return domainerrors.ErrInsufficientStock
	}

	return nil
}
func (u *CartUseCase) AddItem(ctx context.Context, userID int64, variantID int64, quantity int) (*models.Cart, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	if variantID <= 0 {
		return nil, domainerrors.ErrInvalidVariantID
	}
	if quantity <= 0 {
		return nil, domainerrors.ErrInvalidQuantity
	}
	variant, err := u.variantRepo.GetByID(ctx, variantID)
	if err != nil {
		return nil, fmt.Errorf("get product variant: %w", err)
	}
	if err := validateVariant(variant, quantity); err != nil {
		return nil, err
	}
	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	cartID := int64(cart.ID)
	_, err = u.cartRepo.GetCartItemByVariantID(
		ctx,
		cartID,
		variantID,
	)
	if err == nil {
		return nil, domainerrors.ErrCartItemAlreadyExists
	}
	if !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, fmt.Errorf("check existing cart item: %w", err)
	}
	item := &models.CartItem{
		CartID:    cart.ID,
		VariantID: uint(variantID),
		Quantity:  quantity,
	}
	if err := u.cartRepo.CreateCartItem(ctx, item); err != nil {
		existingItem, lookupErr :=
			u.cartRepo.GetCartItemByVariantID(
				ctx,
				cartID,
				variantID,
			)
		if lookupErr == nil && existingItem != nil {
			return nil, domainerrors.ErrCartItemAlreadyExists
		}
		return nil, fmt.Errorf("create cart item: %w", err)
	}
	u.log.Info(
		"cart item added",
		"user_id", userID,
		"variant_id", variantID,
		"quantity", quantity,
	)
	return u.GetCart(ctx, userID)
}
func (u *CartUseCase) UpdateQuantity(ctx context.Context, userID int64, itemID int64, quantity int) (*models.Cart, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	if itemID <= 0 {
		return nil, domainerrors.ErrRecordNotFound
	}
	if quantity <= 0 {
		return nil, domainerrors.ErrInvalidQuantity
	}
	cart, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := u.cartRepo.GetCartItems(
		ctx,
		int64(cart.ID),
	)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	var targetItem *models.CartItem
	for i := range items {
		if int64(items[i].ID) == itemID {
			targetItem = &items[i]
			break
		}
	}
	if targetItem == nil {
		return nil, domainerrors.ErrRecordNotFound
	}
	if err := validateVariant(&targetItem.Variant, quantity); err != nil {
		return nil, err
	}
	if err := u.cartRepo.UpdateCartItemQuantity(
		ctx,
		int64(cart.ID),
		itemID,
		quantity,
	); err != nil {
		return nil, fmt.Errorf("update cart item quantity: %w", err)
	}
	u.log.Info(
		"cart item quantity updated",
		"user_id", userID,
		"item_id", itemID,
		"quantity", quantity,
	)
	return u.GetCart(ctx, userID)
}
func (u *CartUseCase) RemoveItem(ctx context.Context, userID int64, itemID int64) (*models.Cart, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	if itemID <= 0 {
		return nil, domainerrors.ErrRecordNotFound
	}
	cart, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := u.cartRepo.RemoveCartItem(
		ctx,
		int64(cart.ID),
		itemID,
	); err != nil {
		return nil, fmt.Errorf("remove cart item: %w", err)
	}
	u.log.Info(
		"cart item removed",
		"user_id", userID,
		"item_id", itemID,
	)
	return u.GetCart(ctx, userID)
}
func (u *CartUseCase) ClearCart(ctx context.Context,userID int64) (*models.Cart, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}

	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := u.cartRepo.ClearCart(
		ctx,
		int64(cart.ID),
	); err != nil {
		return nil, fmt.Errorf("clear cart: %w", err)
	}

	u.log.Info(
		"cart cleared",
		"user_id", userID,
	)

	return u.GetCart(ctx, userID)
}
