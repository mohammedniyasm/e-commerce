package wishlist

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupWishlistTest(t *testing.T) (
	context.Context,
	*mocks.MockWishlistRepository,
	*WishlistUseCase,
) {
	ctx := context.Background()
	repo := mocks.NewMockWishlistRepository(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	uc := NewWishlistUseCase(repo, logger)
	return ctx, repo, uc
}

func TestCreateWishlist(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)

		repo.EXPECT().CreateWishlist(ctx, mock.MatchedBy(func(w *models.Wishlist) bool {
			return w.UserID == 1 && w.Name == "Favorites"
		})).Return(nil).Once()

		err := uc.CreateWishlist(ctx, 1, "Favorites")
		require.NoError(t, err)
	})

	t.Run("Empty or Whitespace Name", func(t *testing.T) {
		ctx, _, uc := setupWishlistTest(t)

		err := uc.CreateWishlist(ctx, 1, "   ")
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	t.Run("Repo Error", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)
		repoErr := errors.New("database duplicate key error")

		repo.EXPECT().CreateWishlist(ctx, mock.Anything).Return(repoErr).Once()

		err := uc.CreateWishlist(ctx, 1, "Favorites")
		require.ErrorIs(t, err, repoErr)
	})
}

func TestGetWishlists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)
		expected := []models.Wishlist{
			{ID: 1, UserID: 10, Name: "Tech"},
			{ID: 2, UserID: 10, Name: "Books"},
		}

		repo.EXPECT().GetWishlistsByUserID(ctx, uint(10)).Return(expected, nil).Once()

		lists, err := uc.GetWishlists(ctx, 10)
		require.NoError(t, err)
		require.Equal(t, expected, lists)
	})

	t.Run("Repo Error", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)
		repoErr := errors.New("db query error")

		repo.EXPECT().GetWishlistsByUserID(ctx, uint(10)).Return(nil, repoErr).Once()

		lists, err := uc.GetWishlists(ctx, 10)
		require.ErrorIs(t, err, repoErr)
		require.Nil(t, lists)
	})
}

func TestGetWishlistByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)
		expected := &models.Wishlist{ID: 5, UserID: 10, Name: "Gadgets"}

		repo.EXPECT().GetWishlistByID(ctx, uint(10), uint(5)).Return(expected, nil).Once()

		w, err := uc.GetWishlistByID(ctx, 10, 5)
		require.NoError(t, err)
		require.Equal(t, expected, w)
	})

	t.Run("Invalid WishlistID", func(t *testing.T) {
		ctx, _, uc := setupWishlistTest(t)

		w, err := uc.GetWishlistByID(ctx, 10, 0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		require.Nil(t, w)
	})
}

func TestUpdateWishlist(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)

		repo.EXPECT().UpdateWishlist(ctx, uint(5), uint(10), "New Name").Return(nil).Once()

		err := uc.UpdateWishlist(ctx, 5, 10, "New Name")
		require.NoError(t, err)
	})

	t.Run("Invalid Parameters", func(t *testing.T) {
		ctx, _, uc := setupWishlistTest(t)

		err := uc.UpdateWishlist(ctx, 0, 10, "Valid Name")
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		err = uc.UpdateWishlist(ctx, 5, 10, "   ")
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})
}

func TestDeleteWishlist(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)

		repo.EXPECT().DeleteWishlist(ctx, uint(5), uint(10)).Return(nil).Once()

		err := uc.DeleteWishlist(ctx, 5, 10)
		require.NoError(t, err)
	})

	t.Run("Invalid WishlistID", func(t *testing.T) {
		ctx, _, uc := setupWishlistTest(t)

		err := uc.DeleteWishlist(ctx, 0, 10)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})
}

func TestAddProductToWishlist(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)
		expected := &models.Wishlist{ID: 1, UserID: 10, Name: "Default"}

		repo.EXPECT().AddProductToWishlist(ctx, "Default", uint(10), uint(25)).Return(nil).Once()
		repo.EXPECT().GetWishlistByName(ctx, uint(10), "Default").Return(expected, nil).Once()

		res, err := uc.AddProductToWishlist(ctx, "Default", 10, 25)
		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		ctx, _, uc := setupWishlistTest(t)

		res, err := uc.AddProductToWishlist(ctx, "", 10, 25)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		require.Nil(t, res)

		res, err = uc.AddProductToWishlist(ctx, "Default", 10, 0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		require.Nil(t, res)
	})
}

func TestRemoveWishlistItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)
		expected := &models.Wishlist{ID: 1, UserID: 10}

		repo.EXPECT().RemoveWishlistItem(ctx, uint(1), uint(10), uint(3)).Return(nil).Once()
		repo.EXPECT().GetWishlistByID(ctx, uint(1), uint(10)).Return(expected, nil).Once()

		res, err := uc.RemoveWishlistItem(ctx, 1, 10, 3)
		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		ctx, _, uc := setupWishlistTest(t)

		res, err := uc.RemoveWishlistItem(ctx, 0, 10, 3)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		require.Nil(t, res)

		res, err = uc.RemoveWishlistItem(ctx, 1, 10, 0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		require.Nil(t, res)
	})
}

func TestClearWishlist(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)
		expected := &models.Wishlist{ID: 1, UserID: 10}

		repo.EXPECT().ClearWishlist(ctx, uint(1), uint(10)).Return(nil).Once()
		repo.EXPECT().GetWishlistByID(ctx, uint(1), uint(10)).Return(expected, nil).Once()

		res, err := uc.ClearWishlist(ctx, 1, 10)
		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		ctx, _, uc := setupWishlistTest(t)

		res, err := uc.ClearWishlist(ctx, 0, 10)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		require.Nil(t, res)
	})
}

func TestMergeWishlistItems(t *testing.T) {
	t.Run("Success Deduplication and Merge", func(t *testing.T) {
		ctx, repo, uc := setupWishlistTest(t)
		expected := &models.Wishlist{ID: 1, UserID: 10, Name: "Default"}

		// Input contains duplicates: [1, 2, 2, 3] -> unique [1, 2, 3]
		repo.EXPECT().MergeWishlistItems(ctx, "Default", uint(10), []uint{1, 2, 3}).Return(uint(3), uint(0), nil).Once()
		repo.EXPECT().GetWishlistByName(ctx, uint(10), "Default").Return(expected, nil).Once()

		w, accepted, skipped, err := uc.MergeWishlistItems(ctx, "Default", 10, []uint{1, 2, 2, 3})
		require.NoError(t, err)
		require.Equal(t, expected, w)
		require.Equal(t, uint(3), accepted)
		require.Equal(t, uint(1), skipped) // 1 duplicate skipped
	})

	t.Run("Invalid Input", func(t *testing.T) {
		ctx, _, uc := setupWishlistTest(t)

		_, _, _, err := uc.MergeWishlistItems(ctx, "", 10, []uint{1, 2})
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, _, _, err = uc.MergeWishlistItems(ctx, "Default", 10, []uint{})
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, _, _, err = uc.MergeWishlistItems(ctx, "Default", 10, []uint{1, 0})
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})
}
