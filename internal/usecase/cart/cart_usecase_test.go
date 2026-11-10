package cart

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newProductVariant() *models.ProductVariant {
	return &models.ProductVariant{
		ID:           1,
		IsActive:     true,
		Stock:        10,
		SellingPrice: 1000,
		Product: models.Product{
			ID:        1,
			IsActive:  true,
			IsListed:  true,
			DeletedAt: nil,
			Category: models.Category{
				ID:        1,
				IsActive:  true,
				DeletedAt: nil,
			},
			Brand: models.Brand{
				ID:        1,
				IsActive:  true,
				DeletedAt: nil,
			},
		},
	}
}
func TestValidateVariant(t *testing.T) {
	type testCase struct {
		name      string
		variant   *models.ProductVariant
		quantity  int
		wantError error
	}
	tests := []testCase{
		testCase{
			name:      "Valid Variant",
			variant:   newProductVariant(),
			quantity:  5,
			wantError: nil,
		},
		testCase{
			name: "Invalid variant ID",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.ID = 0
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrRecordNotFound,
		},
		testCase{
			name:      "Invalid variant ",
			variant:   nil,
			quantity:  5,
			wantError: domainerrors.ErrRecordNotFound,
		},
		testCase{
			name: "Inactive Variant",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.IsActive = false
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "Invalid variant stock",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.Stock = 0
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "Invalid product",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.Product.ID = 0
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "Inactive product",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.Product.IsActive = false
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "Not listed product",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.Product.IsListed = false
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "deleted product",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				now := time.Now()
				v.Product.DeletedAt = &now
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "deleted category",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				now := time.Now()
				v.Product.Category.DeletedAt = &now
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "deleted brand",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				now := time.Now()
				v.Product.Brand.DeletedAt = &now
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "inactive brand",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.Product.Brand.IsActive = false
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name: "inactive category",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.Product.Category.IsActive = false
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrProductNotAvailable,
		},
		testCase{
			name:      "negative quantity",
			variant:   newProductVariant(),
			quantity:  -5,
			wantError: domainerrors.ErrInvalidQuantity,
		},
		testCase{
			name:      "exceed quantity",
			variant:   newProductVariant(),
			quantity:  6,
			wantError: domainerrors.ErrInvalidQuantity,
		},
		testCase{
			name: "exceed quantity",
			variant: func() *models.ProductVariant {
				v := newProductVariant()
				v.Stock = 3
				return v
			}(),
			quantity:  5,
			wantError: domainerrors.ErrInsufficientStock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVariant(tt.variant, tt.quantity)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("expected: (%v);got: (%v)", tt.wantError, err)
			}
		})
	}
}
func TestGetCart_Success(t *testing.T) {
	ctx := context.Background()
	cartRepo := mocks.NewMockCartRepository(t)
	variantRepo := mocks.NewMockProductVariantRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	usecase := NewCartUseCase(cartRepo, variantRepo, log)
	expectedCart := &models.Cart{
		ID:     1,
		UserID: 10,
	}
	expectedItems := []models.CartItem{
		{
			ID:       1,
			CartID:   1,
			Quantity: 2,
		},
	}
	cartRepo.EXPECT().
		GetCartByUserID(ctx, int64(10)).Return(expectedCart, nil).Once()

	cartRepo.EXPECT().
		GetCartItems(ctx, int64(1)).Return(expectedItems, nil).Once()

	result, err := usecase.GetCart(ctx, int64(10))
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, expectedCart.ID, result.ID)
	require.Equal(t, expectedItems, result.Items)

}
func TestGerCart_CartRepositoryError(t *testing.T) {
	ctx := context.Background()
	cartRepo := mocks.NewMockCartRepository(t)
	variantRepo := mocks.NewMockProductVariantRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	usecase := NewCartUseCase(cartRepo, variantRepo, log)
	expectedError := errors.New("database connection failed")

	cartRepo.EXPECT().
		GetCartByUserID(ctx, int64(10)).
		Return(nil, expectedError).Once()

	result, err := usecase.GetCart(ctx, 10)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, expectedError)
}
func TestAddItem_Success(t *testing.T) {
	ctx := context.Background()
	cartRepo := mocks.NewMockCartRepository(t)
	variantRepo := mocks.NewMockProductVariantRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	usecase := NewCartUseCase(cartRepo, variantRepo, log)

	expectedCart := &models.Cart{
		ID:     1,
		UserID: 10,
	}

	expectedItems := []models.CartItem{
		{
			ID:        1,
			CartID:    1,
			VariantID: 1,
			Quantity:  2,
		},
	}
	variant := newProductVariant()
	variantRepo.EXPECT().
		GetByID(ctx, int64(1)).
		Return(variant, nil).
		Once()

	cartRepo.EXPECT().
		GetCartByUserID(ctx, int64(10)).
		Return(expectedCart, nil).
		Twice()

	cartRepo.EXPECT().
		GetCartItemByVariantID(ctx, int64(1), int64(1)).
		Return(nil, domainerrors.ErrRecordNotFound).
		Once()

	cartRepo.EXPECT().
		CreateCartItem(ctx, mock.Anything).
		Return(nil).
		Once()

	cartRepo.EXPECT().
		GetCartItems(ctx, int64(1)).
		Return(expectedItems, nil).
		Once()

	result, err := usecase.AddItem(ctx, 10, 1, 2)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, expectedCart.ID, result.ID)
	require.Equal(t, expectedItems, result.Items)
}
func TestAddItem_InvalidInputs(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		variantID int64
		quantity  int
		wantError error
	}{
		{
			name:      "invalid user ID",
			userID:    0,
			variantID: 1,
			quantity:  2,
			wantError: domainerrors.ErrInvalidUserID,
		},
		{
			name:      "invalid variant ID",
			userID:    10,
			variantID: 0,
			quantity:  2,
			wantError: domainerrors.ErrInvalidVariantID,
		},
		{
			name:      "invalid quantity",
			userID:    10,
			variantID: 1,
			quantity:  0,
			wantError: domainerrors.ErrInvalidQuantity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := mocks.NewMockCartRepository(t)
			variantRepo := mocks.NewMockProductVariantRepository(t)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			usecase := NewCartUseCase(cartRepo, variantRepo, log)

			result, err := usecase.AddItem(
				context.Background(),
				tt.userID,
				tt.variantID,
				tt.quantity,
			)

			require.ErrorIs(t, err, tt.wantError)
			require.Nil(t, result)
		})
	}
}
func TestAddItem_GetVariantError(t *testing.T) {
	ctx := context.Background()

	cartRepo := mocks.NewMockCartRepository(t)
	variantRepo := mocks.NewMockProductVariantRepository(t)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	usecase := NewCartUseCase(cartRepo, variantRepo, log)

	dbErr := errors.New("database connection failed")

	variantRepo.EXPECT().
		GetByID(ctx, int64(1)).
		Return(nil, dbErr).
		Once()

	result, err := usecase.AddItem(ctx, 10, 1, 2)

	require.ErrorIs(t, err, dbErr)
	require.Nil(t, result)
}
func TestAddItem_VariantUnavailable(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*models.ProductVariant)
	}{
		{
			name: "inactive variant",
			modify: func(v *models.ProductVariant) {
				v.IsActive = false
			},
		},
		{
			name: "variant has no stock",
			modify: func(v *models.ProductVariant) {
				v.Stock = 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cartRepo := mocks.NewMockCartRepository(t)
			variantRepo := mocks.NewMockProductVariantRepository(t)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			usecase := NewCartUseCase(cartRepo, variantRepo, log)

			variant := newProductVariant()
			tt.modify(variant)

			variantRepo.EXPECT().
				GetByID(ctx, int64(1)).
				Return(variant, nil).
				Once()

			result, err := usecase.AddItem(ctx, 10, 1, 2)

			require.ErrorIs(t, err, domainerrors.ErrProductNotAvailable)
			require.Nil(t, result)
		})
	}
}
func TestAddItem_GetCartError(t *testing.T) {
	ctx := context.Background()

	cartRepo := mocks.NewMockCartRepository(t)
	variantRepo := mocks.NewMockProductVariantRepository(t)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	usecase := NewCartUseCase(cartRepo, variantRepo, log)

	dbErr := errors.New("database connection failed")

	variant := newProductVariant()

	variantRepo.EXPECT().
		GetByID(ctx, int64(1)).
		Return(variant, nil).
		Once()

	cartRepo.EXPECT().
		GetCartByUserID(ctx, int64(10)).
		Return(nil, dbErr).
		Once()

	result, err := usecase.AddItem(ctx, 10, 1, 2)

	require.ErrorIs(t, err, dbErr)
	require.Nil(t, result)
}
