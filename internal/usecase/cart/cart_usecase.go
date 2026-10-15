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
		u.log.Error("failed to retrieve cart",
			"user_id", userID,
			"error", err,
		)
		return nil, fmt.Errorf("get cart: %w", err)
	}

	newCart := &models.Cart{
		UserID: uint(userID),
	}

	if err := u.cartRepo.CreateCart(ctx, newCart); err != nil {
		u.log.Error("failed to create cart",
			"user_id", userID,
			"error", err,
		)
		return nil, fmt.Errorf("create cart: %w", err)
	}
	u.log.Info("cart created successfully",
		"user_id", userID,
		"cart_id", newCart.ID,
	)
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
		u.log.Error("failed to retrieve cart items",
			"user_id", userID,
			"cart_id", cart.ID,
			"error", err,
		)
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	cart.Items = items
	u.log.Debug("cart retrieved with items",
		"user_id", userID,
		"cart_id", cart.ID,
		"item_count", len(items),
	)
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
		variant.Product.DeletedAt != nil ||
		!variant.Product.Category.IsActive ||
		variant.Product.Category.DeletedAt != nil ||
		!variant.Product.Brand.IsActive ||
		variant.Product.Brand.DeletedAt != nil {
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
		if !errors.Is(err, domainerrors.ErrRecordNotFound) {
			u.log.Error("failed to retrieve product variant",
				"user_id", userID,
				"variant_id", variantID,
				"error", err,
			)
		}
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
		ctx, cartID, variantID,
	)
	if err == nil {
		return nil, domainerrors.ErrCartItemAlreadyExists
	}
	if !errors.Is(err, domainerrors.ErrRecordNotFound) {
		u.log.Error("failed to check existing cart item",
			"user_id", userID,
			"cart_id", cartID,
			"variant_id", variantID,
			"error", err,
		)
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
				ctx, cartID, variantID,
			)

		if lookupErr == nil && existingItem != nil {
			return nil, domainerrors.ErrCartItemAlreadyExists
		}

		u.log.Error("failed to create cart item",
			"user_id", userID,
			"cart_id", cartID,
			"variant_id", variantID,
			"quantity", quantity,
			"error", err,
		)
		return nil, fmt.Errorf("create cart item: %w", err)
	}

	u.log.Info("cart item added successfully",
		"user_id", userID,
		"cart_id", cartID,
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
		if !errors.Is(err, domainerrors.ErrRecordNotFound) {
			u.log.Error("failed to retrieve cart for quantity update",
				"user_id", userID,
				"item_id", itemID,
				"error", err,
			)
		}
		return nil, err
	}

	items, err := u.cartRepo.GetCartItems(ctx, int64(cart.ID))
	if err != nil {
		u.log.Error("failed to retrieve cart items for quantity update",
			"user_id", userID,
			"cart_id", cart.ID,
			"item_id", itemID,
			"error", err,
		)
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
		ctx, int64(cart.ID), itemID, quantity,
	); err != nil {
		u.log.Error("failed to update cart item quantity",
			"user_id", userID,
			"cart_id", cart.ID,
			"item_id", itemID,
			"quantity", quantity,
			"error", err,
		)
		return nil, fmt.Errorf("update cart item quantity: %w", err)
	}
	u.log.Info("cart item quantity updated",
		"user_id", userID,
		"cart_id", cart.ID,
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
		if !errors.Is(err, domainerrors.ErrRecordNotFound) {
			u.log.Error("failed to retrieve cart for item removal",
				"user_id", userID,
				"item_id", itemID,
				"error", err,
			)
		}
		return nil, err
	}

	if err := u.cartRepo.RemoveCartItem(
		ctx, int64(cart.ID), itemID,
	); err != nil {
		if !errors.Is(err, domainerrors.ErrRecordNotFound) {
			u.log.Error("failed to remove cart item",
				"user_id", userID,
				"cart_id", cart.ID,
				"item_id", itemID,
				"error", err,
			)
		}
		return nil, fmt.Errorf("remove cart item: %w", err)
	}

	u.log.Info("cart item removed successfully",
		"user_id", userID,
		"cart_id", cart.ID,
		"item_id", itemID,
	)

	return u.GetCart(ctx, userID)
}
func (u *CartUseCase) ClearCart(ctx context.Context, userID int64) (*models.Cart, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}

	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := u.cartRepo.ClearCart(ctx, int64(cart.ID)); err != nil {
		u.log.Error("failed to clear cart",
			"user_id", userID,
			"cart_id", cart.ID,
			"error", err,
		)
		return nil, fmt.Errorf("clear cart: %w", err)
	}

	u.log.Info("cart cleared successfully",
		"user_id", userID,
		"cart_id", cart.ID,
	)

	return u.GetCart(ctx, userID)
}
func (u *CartUseCase) ValidateGuestCart(ctx context.Context, items []interfaces.GuestCartItemInput) ([]interfaces.GuestCartItemResult, error) {
	results := make(
		[]interfaces.GuestCartItemResult,
		0,
		len(items),
	)

	if len(items) == 0 {
		return results, nil
	}

	variantIDs := make([]int64, 0, len(items))
	for _, item := range items {
		if item.VariantID <= 0 {
			return nil, domainerrors.ErrInvalidVariantID
		}
		if item.Quantity <= 0 {
			return nil, domainerrors.ErrInvalidQuantity
		}
		variantIDs = append(variantIDs, item.VariantID)
	}

	variants, err := u.variantRepo.GetByIDs(ctx, variantIDs)
	if err != nil {
		u.log.Error("failed to retrieve variants for guest cart validation",
			"item_count", len(items),
			"error", err,
		)
		return nil, fmt.Errorf(
			"failed to retrieve variants for guest cart validation: %w",
			err,
		)
	}

	variantMap := make(map[int64]models.ProductVariant, len(variants))
	for _, variant := range variants {
		variantMap[int64(variant.ID)] = variant
	}

	for _, item := range items {
		result := interfaces.GuestCartItemResult{
			VariantID:         item.VariantID,
			RequestedQuantity: item.Quantity,
			IsAvailable:       false,
		}
		variant, exists := variantMap[item.VariantID]
		if !exists {
			result.AvailabilityReason = "variant_not_found"
			results = append(results, result)
			continue
		}

		result.AvailableStock = variant.Stock
		result.CurrentPrice = variant.SellingPrice

		switch {
		case variant.Product.ID == 0 ||
			!variant.Product.IsActive ||
			!variant.Product.IsListed ||
			variant.Product.DeletedAt != nil ||
			!variant.Product.Category.IsActive ||
			!variant.Product.Brand.IsActive ||
			variant.Product.Category.DeletedAt != nil ||
			variant.Product.Brand.DeletedAt != nil:
			result.AvailabilityReason = "product_unavailable"

		case !variant.IsActive:
			result.AvailabilityReason = "variant_inactive"

		case variant.Stock <= 0:
			result.AvailabilityReason = "out_of_stock"

		case item.Quantity > variant.Stock:
			result.AvailabilityReason = "insufficient_stock"

		default:
			result.IsAvailable = true
			result.AvailabilityReason = ""
		}

		results = append(results, result)
	}

	u.log.Debug("guest cart validation completed",
		"item_count", len(items),
		"available_count", countAvailableGuestItems(results),
	)

	return results, nil
}
func countAvailableGuestItems(items []interfaces.GuestCartItemResult) int {
	count := 0
	for _, item := range items {
		if item.IsAvailable {
			count++
		}
	}
	return count
}
func (u *CartUseCase) MergeGuestCart(ctx context.Context, userID int64, items []interfaces.GuestCartItemInput) (*models.Cart, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}

	if len(items) == 0 {
		u.log.Debug("guest cart is empty; returning user cart",
			"user_id", userID,
		)
		return u.GetCart(ctx, userID)
	}

	for _, item := range items {
		if item.VariantID <= 0 {
			return nil, domainerrors.ErrInvalidVariantID
		}
		if item.Quantity <= 0 {
			return nil, domainerrors.ErrInvalidQuantity
		}
	}

	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	cartID := int64(cart.ID)

	err = u.cartRepo.WithTransaction(
		ctx,
		func(txRepo interfaces.CartRepository) error {
			for _, item := range items {
				variant, err := u.variantRepo.GetByID(
					ctx, item.VariantID,
				)
				if err != nil {
					return fmt.Errorf("get product variant: %w", err)
				}
				existingItem, err := txRepo.GetCartItemByVariantID(
					ctx, cartID, item.VariantID,
				)
				if err == nil {
					newQuantity := existingItem.Quantity + item.Quantity

					if err := validateVariant(variant, newQuantity); err != nil {
						return err
					}

					if err := txRepo.UpdateCartItemQuantity(
						ctx, cartID, int64(existingItem.ID), newQuantity,
					); err != nil {
						return fmt.Errorf("update cart item: %w", err)
					}
					continue
				}

				if !errors.Is(err, domainerrors.ErrRecordNotFound) {
					return fmt.Errorf("get existing cart item: %w", err)
				}

				if err := validateVariant(variant, item.Quantity); err != nil {
					return err
				}

				newItem := &models.CartItem{
					CartID:    cart.ID,
					VariantID: uint(item.VariantID),
					Quantity:  item.Quantity,
				}

				if err := txRepo.CreateCartItem(ctx, newItem); err != nil {
					return fmt.Errorf("create cart item: %w", err)
				}
			}
			return nil
		},
	)

	if err != nil {
		u.log.Error("failed to merge guest cart",
			"user_id", userID,
			"cart_id", cartID,
			"item_count", len(items),
			"error", err,
		)
		return nil, fmt.Errorf("merge guest cart: %w", err)
	}

	u.log.Info("guest cart merged successfully",
		"user_id", userID,
		"cart_id", cartID,
		"item_count", len(items),
	)

	return u.GetCart(ctx, userID)
}
