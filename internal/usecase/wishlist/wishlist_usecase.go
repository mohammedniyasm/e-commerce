package wishlist

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"log/slog"
	"strings"
)

type WishlistUseCase struct {
	wishlistRepo interfaces.WishlistRepository
	log          *slog.Logger
}

func NewWishlistUseCase(
	wishlistRepo interfaces.WishlistRepository,
	log *slog.Logger,
) *WishlistUseCase {
	return &WishlistUseCase{
		wishlistRepo: wishlistRepo,
		log:          log,
	}
}
func (u *WishlistUseCase) CreateWishlist(ctx context.Context, userID uint, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return domainerrors.ErrInvalidInput
	}
	wishlist := &models.Wishlist{UserID: userID, Name: name}
	if err := u.wishlistRepo.CreateWishlist(ctx, wishlist); err != nil {
		u.log.ErrorContext(ctx,
			"failed to create wishlist",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	return nil
}
func (u *WishlistUseCase) GetWishlists(ctx context.Context, userID uint) ([]models.Wishlist, error) {
	wishlists, err := u.wishlistRepo.GetWishlistsByUserID(ctx, userID)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to retrieve user wishlists",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}
	return wishlists, nil
}
func (u *WishlistUseCase) GetWishlistByID(ctx context.Context, userID uint, wishlistID uint) (*models.Wishlist, error) {
	if wishlistID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	wishlist, err := u.wishlistRepo.GetWishlistByID(
		ctx,
		userID,
		wishlistID,
	)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to retrieve wishlist",
			"wishlist_id", wishlistID,
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}
	return wishlist, nil
}
func (u *WishlistUseCase) UpdateWishlist(ctx context.Context, wishlistID uint, userID uint, name string) error {
	name = strings.TrimSpace(name)
	if wishlistID == 0 || name == "" {
		return domainerrors.ErrInvalidInput
	}
	err := u.wishlistRepo.UpdateWishlist(
		ctx,
		wishlistID,
		userID,
		name,
	)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to update wishlist",
			"wishlist_id", wishlistID,
			"user_id", userID,
			"error", err,
		)
		return err
	}
	return nil
}
func (u *WishlistUseCase) DeleteWishlist(ctx context.Context, wishlistID uint, userID uint) error {
	if wishlistID == 0 {
		return domainerrors.ErrInvalidInput
	}
	err := u.wishlistRepo.DeleteWishlist(
		ctx,
		wishlistID,
		userID,
	)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to delete wishlist",
			"wishlist_id", wishlistID,
			"user_id", userID,
			"error", err,
		)
		return err
	}
	return nil
}
func (u *WishlistUseCase) AddProductToWishlist(ctx context.Context, wishlistName string, userID uint, productID uint) (*models.Wishlist, error) {
	wishlistName = strings.TrimSpace(wishlistName)
	if wishlistName == "" || productID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	err := u.wishlistRepo.AddProductToWishlist(
		ctx,
		wishlistName,
		userID,
		productID,
	)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to add product to wishlist",
			"user_id", userID,
			"product_id", productID,
			"wishlist_name", wishlistName,
			"error", err,
		)
		return nil, err
	}
	return u.wishlistRepo.GetWishlistByName(ctx, userID, wishlistName)
}
func (u *WishlistUseCase) RemoveWishlistItem(ctx context.Context, wishlistID uint, userID uint, itemID uint) (*models.Wishlist, error) {
	if wishlistID == 0 || itemID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	err := u.wishlistRepo.RemoveWishlistItem(
		ctx,
		wishlistID,
		userID,
		itemID,
	)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to remove wishlist item",
			"wishlist_id", wishlistID,
			"item_id", itemID,
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}
	return u.wishlistRepo.GetWishlistByID(ctx, wishlistID, userID)
}
func (u *WishlistUseCase) ClearWishlist(ctx context.Context, wishlistID uint, userID uint) (*models.Wishlist, error) {
	if wishlistID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	err := u.wishlistRepo.ClearWishlist(ctx, wishlistID, userID)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to clear wishlist",
			"wishlist_id", wishlistID,
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}
	return u.wishlistRepo.GetWishlistByID(ctx, wishlistID, userID)
}
func (u *WishlistUseCase) MergeWishlistItems(ctx context.Context, wishlistName string, userID uint, productIDs []uint) (*models.Wishlist, uint, uint, error) {
	wishlistName = strings.TrimSpace(wishlistName)
	if wishlistName == "" || len(productIDs) == 0 {
		return nil, 0, 0, domainerrors.ErrInvalidInput
	}
	// Remove duplicate product IDs.
	uniqueIDs := make([]uint, 0, len(productIDs))
	seen := make(map[uint]struct{}, len(productIDs))
	for _, id := range productIDs {
		if id == 0 {
			return nil, 0, 0, domainerrors.ErrInvalidInput
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	accepted, skipped, err := u.wishlistRepo.MergeWishlistItems(
		ctx,
		wishlistName,
		userID,
		uniqueIDs,
	)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to merge wishlist items",
			"user_id", userID,
			"error", err,
		)
		return nil, 0, 0, err
	}
	skipped += uint(len(productIDs) - len(uniqueIDs))
	wishlist, err := u.wishlistRepo.GetWishlistByName(ctx, userID, wishlistName)
	if err != nil {
		return nil, 0, 0, err
	}
	return wishlist, accepted, skipped, nil
}

var _ interfaces.WishlistUseCase = (*WishlistUseCase)(nil)
