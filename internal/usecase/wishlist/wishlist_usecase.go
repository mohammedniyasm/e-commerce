package wishlist

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"log/slog"
)

type WishlistUseCase struct {
	wishlistRepo interfaces.WishlistRepository
	productRepo  interfaces.ProductRepository
	log          *slog.Logger
}

func NewWishlistUseCase(
	wishlistRepo interfaces.WishlistRepository,
	productRepo interfaces.ProductRepository,
	log *slog.Logger,
) *WishlistUseCase {
	return &WishlistUseCase{
		wishlistRepo: wishlistRepo,
		productRepo:  productRepo,
		log:          log,
	}
}
func (u *WishlistUseCase) getOrCreateWishlist(ctx context.Context, userID int64) (*models.Wishlist, error) {
	wishlist, err := u.wishlistRepo.GetWishlistByUserID(ctx, userID)
	if err == nil {
		return wishlist, nil
	}
	if !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, err
	}
	newWishlist := &models.Wishlist{
		UserID: uint(userID),
	}
	if err := u.wishlistRepo.CreateWishlist(ctx, newWishlist); err != nil {
		existingWishlist, lookupErr :=
			u.wishlistRepo.GetWishlistByUserID(ctx, userID)
		if lookupErr == nil {
			return existingWishlist, nil
		}
		return nil, err
	}
	return newWishlist, nil
}
func (u *WishlistUseCase) ValidateProduct(ctx context.Context, id int64) error {
	product, err := u.productRepo.GetByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			return domainerrors.ErrRecordNotFound
		}
		return err
	}
	if !product.IsActive ||
		!product.IsListed ||
		product.DeletedAt != nil ||
		product.Category.DeletedAt != nil ||
		product.Brand.DeletedAt != nil ||
		!product.Category.IsActive ||
		!product.Brand.IsActive {
		return domainerrors.ErrProductNotAvailable
	}
	return nil
}
func (u *WishlistUseCase) GetWishlist(ctx context.Context, userID int64) (*models.Wishlist, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	wishlist, err := u.getOrCreateWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := u.wishlistRepo.GetWishlistItems(ctx, int64(wishlist.ID))
	if err != nil {
		return nil, err
	}
	wishlist.Items = items
	return wishlist, nil
}
func (u *WishlistUseCase) AddProduct(ctx context.Context, userID int64, productID int64) (*models.Wishlist, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	if productID <= 0 {
		return nil, domainerrors.ErrInvalidProductID
	}
	if err := u.ValidateProduct(ctx, productID); err != nil {
		return nil, err
	}
	wishlist, err := u.getOrCreateWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}
	_, err = u.wishlistRepo.GetWishlistItemByProductID(ctx, int64(wishlist.ID), productID)
	if err == nil {
		return nil, domainerrors.ErrWishlistItemAlreadyExists
	}
	if !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, err
	}
	item := &models.WishlistItem{
		WishlistID: wishlist.ID,
		ProductID:  uint(productID),
	}
	if err := u.wishlistRepo.CreateWishlistItem(ctx, item); err != nil {
		existingItem, lookupErr := u.wishlistRepo.GetWishlistItemByProductID(ctx, int64(wishlist.ID), productID)
		if lookupErr == nil && existingItem != nil {
			return nil, domainerrors.ErrWishlistItemAlreadyExists
		}
		return nil, err
	}
	u.log.Info(
		"product added to wishlist",
		"user_id", userID,
		"product_id", productID,
	)
	return u.GetWishlist(ctx, userID)
}
func (u *WishlistUseCase) RemoveItem(ctx context.Context, userID int64, itemID int64) (*models.Wishlist, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	if itemID <= 0 {
		return nil, domainerrors.ErrInvalidProductID
	}
	wishlist, err := u.wishlistRepo.GetWishlistByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := u.ValidateProduct(ctx, itemID); err != nil {
		return nil, err
	}
	_, err = u.wishlistRepo.GetWishlistItemByProductID(ctx, int64(wishlist.ID), itemID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			return nil, domainerrors.ErrProductNotFound
		}
		return nil, err
	}
	if err := u.wishlistRepo.RemoveWishlistItem(ctx, int64(wishlist.ID), itemID); err != nil {
		return nil, err
	}
	u.log.Info(
		"product removed from wishlist",
		"user_id", userID,
		"item_id", itemID,
	)
	return u.GetWishlist(ctx, userID)
}
func (u *WishlistUseCase) ClearWishlist(ctx context.Context, userID int64) (*models.Wishlist, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	wishlist, err := u.getOrCreateWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := u.wishlistRepo.ClearWishlist(ctx, int64(wishlist.ID)); err != nil {
		return nil, err
	}
	u.log.Info(
		"wishlist cleared",
		"user_id", userID,
	)
	return u.GetWishlist(ctx, userID)
}
func (u *WishlistUseCase) MergeGuestWishlist(ctx context.Context, userID int64, productIDs []int64) (*models.Wishlist, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrRecordNotFound
	}
	if len(productIDs) == 0 {
		return u.GetWishlist(ctx, userID)
	}
	uniqueProductIDs := make([]int64, 0, len(productIDs))
	seen := make(map[int64]struct{}, len(productIDs))
	for _, productID := range productIDs {
		if productID <= 0 {
			return nil, domainerrors.ErrInvalidProductID
		}
		if _, exists := seen[productID]; exists {
			continue
		}
		seen[productID] = struct{}{}
		uniqueProductIDs = append(uniqueProductIDs, productID)
	}
	for _, productID := range uniqueProductIDs {
		if err := u.ValidateProduct(ctx, productID); err != nil {
			return nil, err

		}
	}
	wishlist, err := u.getOrCreateWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}
	err = u.wishlistRepo.WithTransaction(ctx, func(txRepo interfaces.WishlistRepository) error {
		for _, productID := range uniqueProductIDs {
			_, err := txRepo.GetWishlistItemByProductID(ctx, int64(wishlist.ID), productID)
			if err == nil {
				continue
			}
			if !errors.Is(err, domainerrors.ErrRecordNotFound) {
				return err
			}
			item := &models.WishlistItem{
				WishlistID: wishlist.ID,
				ProductID:  uint(productID),
			}
			if err := txRepo.CreateWishlistItem(ctx, item); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	u.log.Info(
		"guest wishlist merged",
		"user_id", userID,
		"product_count", len(uniqueProductIDs),
	)
	return u.GetWishlist(ctx, userID)
}
