package productvariant

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	domainErrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type variantFixture struct {
	uc          *ProductVariantUseCase
	variantRepo *mocks.MockProductVariantRepository
	productRepo *mocks.MockProductRepository
	rateLimiter *mocks.MockRateLimiter
}

func newVariantFixture(t *testing.T) *variantFixture {
	t.Helper()
	variantRepo := mocks.NewMockProductVariantRepository(t)
	productRepo := mocks.NewMockProductRepository(t)
	rateLimiter := mocks.NewMockRateLimiter(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &variantFixture{
		uc:          NewProductVariantUseCase(variantRepo, productRepo, rateLimiter, log),
		variantRepo: variantRepo,
		productRepo: productRepo,
		rateLimiter: rateLimiter,
	}
}

func TestProductVariantUseCase_CreateVariant(t *testing.T) {
	ctx := context.Background()
	size := " M "
	color := " Blue "

	t.Run("validation errors", func(t *testing.T) {
		f := newVariantFixture(t)
		_, err := f.uc.CreateVariant(ctx, 0, &size, &color, "SKU-1", 100, 90, 5)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductID)

		f.productRepo.EXPECT().GetByID(ctx, uint(1)).Return(&models.Product{ID: 1}, nil).Times(5)

		_, err = f.uc.CreateVariant(ctx, 1, &size, &color, "   ", 100, 90, 5)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidVariantSKU)

		_, err = f.uc.CreateVariant(ctx, 1, &size, &color, "SKU-1", 0, 90, 5)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidMRP)

		_, err = f.uc.CreateVariant(ctx, 1, &size, &color, "SKU-1", 100, 0, 5)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidSellingPrice)

		_, err = f.uc.CreateVariant(ctx, 1, &size, &color, "SKU-1", 100, 120, 5)
		assert.ErrorIs(t, err, domainErrors.ErrSellingPriceGreaterThanMRP)

		_, err = f.uc.CreateVariant(ctx, 1, &size, &color, "SKU-1", 100, 90, -1)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidStock)
	})

	t.Run("duplicate sku or attributes", func(t *testing.T) {
		f := newVariantFixture(t)
		f.productRepo.EXPECT().GetByID(ctx, uint(1)).Return(&models.Product{ID: 1}, nil).Twice()

		f.variantRepo.EXPECT().GetBySKU(ctx, "SKU-1").Return(&models.ProductVariant{ID: 10}, nil).Once()
		_, err := f.uc.CreateVariant(ctx, 1, &size, &color, "sku-1", 100, 90, 5)
		assert.ErrorIs(t, err, domainErrors.ErrVariantSKUAlreadyExists)

		normSize, normColor := "M", "Blue"
		f.variantRepo.EXPECT().GetBySKU(ctx, "SKU-1").Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.variantRepo.EXPECT().GetByProductAndAttributes(ctx, int64(1), &normSize, &normColor).
			Return(&models.ProductVariant{ID: 11}, nil).Once()
		_, err = f.uc.CreateVariant(ctx, 1, &size, &color, "sku-1", 100, 90, 5)
		assert.ErrorIs(t, err, domainErrors.ErrVariantAlreadyExists)
	})

	t.Run("success", func(t *testing.T) {
		f := newVariantFixture(t)
		normSize, normColor := "M", "Blue"

		f.productRepo.EXPECT().GetByID(ctx, uint(1)).Return(&models.Product{ID: 1}, nil).Once()
		f.variantRepo.EXPECT().GetBySKU(ctx, "SKU-1").Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.variantRepo.EXPECT().GetByProductAndAttributes(ctx, int64(1), &normSize, &normColor).
			Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.variantRepo.EXPECT().Create(ctx, mock.MatchedBy(func(v *models.ProductVariant) bool {
			return v.ProductID == 1 && v.SKU == "SKU-1" &&
				v.Size != nil && *v.Size == "M" &&
				v.Color != nil && *v.Color == "Blue" &&
				v.MRP == 100 && v.SellingPrice == 80 && v.Stock == 12
		})).Return(nil).Once()

		got, err := f.uc.CreateVariant(ctx, 1, &size, &color, " sku-1 ", 100, 80, 12)
		require.NoError(t, err)
		assert.Equal(t, "SKU-1", got.SKU)
	})
}

func TestProductVariantUseCase_GetUpdateDeleteListStock(t *testing.T) {
	ctx := context.Background()

	t.Run("GetVariant", func(t *testing.T) {
		f := newVariantFixture(t)
		_, err := f.uc.GetVariant(ctx, 0)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidVariantID)

		expected := &models.ProductVariant{ID: 3}
		f.variantRepo.EXPECT().GetByID(ctx, int64(3)).Return(expected, nil).Once()
		got, err := f.uc.GetVariant(ctx, 3)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("UpdateVariant", func(t *testing.T) {
		f := newVariantFixture(t)
		_, err := f.uc.UpdateVariant(ctx, 0, nil, nil, "SKU-2", 200, 150, 5, nil)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidVariantID)

		oldSize := "S"
		existing := &models.ProductVariant{ID: 4, ProductID: 1, SKU: "SKU-OLD", Size: &oldSize, IsActive: true}
		newSize := "L"
		active := false

		f.variantRepo.EXPECT().GetByID(ctx, int64(4)).Return(existing, nil).Once()
		f.variantRepo.EXPECT().GetBySKU(ctx, "SKU-NEW").Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.variantRepo.EXPECT().GetByProductAndAttributes(ctx, int64(1), &newSize, (*string)(nil)).
			Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.variantRepo.EXPECT().Update(ctx, mock.MatchedBy(func(v *models.ProductVariant) bool {
			return v.SKU == "SKU-NEW" && v.SellingPrice == 150 && v.Stock == 20 && !v.IsActive
		})).Return(nil).Once()

		got, err := f.uc.UpdateVariant(ctx, 4, &newSize, nil, "sku-new", 200, 150, 20, &active)
		require.NoError(t, err)
		assert.Equal(t, "SKU-NEW", got.SKU)
		assert.False(t, got.IsActive)
	})

	t.Run("DeleteVariant", func(t *testing.T) {
		f := newVariantFixture(t)
		assert.ErrorIs(t, f.uc.DeleteVariant(ctx, 0), domainErrors.ErrInvalidVariantID)

		f.variantRepo.EXPECT().GetByID(ctx, int64(6)).Return(&models.ProductVariant{ID: 6}, nil).Once()
		f.variantRepo.EXPECT().Delete(ctx, int64(6)).Return(nil).Once()
		require.NoError(t, f.uc.DeleteVariant(ctx, 6))
	})

	t.Run("ListVariantsByProductID", func(t *testing.T) {
		f := newVariantFixture(t)
		_, err := f.uc.ListVariantsByProductID(ctx, 0)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductID)

		expected := []models.ProductVariant{{ID: 1}}
		f.productRepo.EXPECT().GetByID(ctx, uint(2)).Return(&models.Product{ID: 2}, nil).Once()
		f.variantRepo.EXPECT().ListByProductID(ctx, int64(2)).Return(expected, nil).Once()

		got, err := f.uc.ListVariantsByProductID(ctx, 2)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("IncreaseStock", func(t *testing.T) {
		f := newVariantFixture(t)
		assert.Error(t, f.uc.IncreaseStock(ctx, 0, 5))
		assert.Error(t, f.uc.IncreaseStock(ctx, 3, 0))

		dbErr := errors.New("db error")
		f.variantRepo.EXPECT().IncreaseStock(ctx, int64(3), 5).Return(dbErr).Once()
		assert.ErrorIs(t, f.uc.IncreaseStock(ctx, 3, 5), dbErr)

		f.variantRepo.EXPECT().IncreaseStock(ctx, int64(3), 5).Return(nil).Once()
		require.NoError(t, f.uc.IncreaseStock(ctx, 3, 5))
	})
}
