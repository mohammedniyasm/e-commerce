package products

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

type productFixture struct {
	uc           *ProductUseCase
	productRepo  *mocks.MockProductRepository
	categoryRepo *mocks.MockCategoryRepository
	brandRepo    *mocks.MockBrandRepository
}

func newProductFixture(t *testing.T) *productFixture {
	t.Helper()
	productRepo := mocks.NewMockProductRepository(t)
	categoryRepo := mocks.NewMockCategoryRepository(t)
	brandRepo := mocks.NewMockBrandRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &productFixture{
		uc:           NewProductUseCase(productRepo, categoryRepo, brandRepo, log),
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
	}
}

func TestProductHelpers_Slugify(t *testing.T) {
	assert.Equal(t, "classic-cotton-t-shirt", slugify("  Classic Cotton T-Shirt! "))
	assert.Equal(t, "hello-world-2025", slugify("Hello---World_2025"))
	assert.Empty(t, slugify("!!!"))
}

func TestProductUseCase_CreateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("validation errors", func(t *testing.T) {
		f := newProductFixture(t)

		_, err := f.uc.CreateProduct(ctx, "   ", nil, nil, 1, 1)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductName)

		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Category{ID: 1, IsActive: true}, nil).Once()
		f.brandRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Brand{ID: 1, IsActive: true}, nil).Once()
		_, err = f.uc.CreateProduct(ctx, "!!!", nil, nil, 1, 1)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductName)

		_, err = f.uc.CreateProduct(ctx, "Cotton Shirt", nil, nil, 0, 1)
		assert.Error(t, err)

		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Category{ID: 1, IsActive: true}, nil).Once()
		_, err = f.uc.CreateProduct(ctx, "Cotton Shirt", nil, nil, 1, 0)
		assert.Error(t, err)
	})

	t.Run("inactive or missing category and brand", func(t *testing.T) {
		f := newProductFixture(t)
		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(nil, domainErrors.ErrRecordNotFound).Once()
		_, err := f.uc.CreateProduct(ctx, "Cotton Shirt", nil, nil, 1, 2)
		assert.ErrorIs(t, err, domainErrors.ErrRecordNotFound)

		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Category{ID: 1, IsActive: false}, nil).Once()
		_, err = f.uc.CreateProduct(ctx, "Cotton Shirt", nil, nil, 1, 2)
		assert.ErrorIs(t, err, domainErrors.ErrCategoryInactive)

		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Category{ID: 1, IsActive: true}, nil).Once()
		f.brandRepo.EXPECT().GetByID(ctx, int64(2)).Return(&models.Brand{ID: 2, IsActive: false}, nil).Once()
		_, err = f.uc.CreateProduct(ctx, "Cotton Shirt", nil, nil, 1, 2)
		assert.ErrorIs(t, err, domainErrors.ErrBrandInactive)
	})

	t.Run("success resolves slug collision and creates product", func(t *testing.T) {
		f := newProductFixture(t)
		shortDesc := "  Soft cotton  "
		desc := "  Full description  "

		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Category{ID: 1, IsActive: true}, nil).Once()
		f.brandRepo.EXPECT().GetByID(ctx, int64(2)).Return(&models.Brand{ID: 2, IsActive: true}, nil).Once()
		f.productRepo.EXPECT().GetBySlug(ctx, "cotton-shirt").Return(&models.Product{ID: 99, Slug: "cotton-shirt"}, nil).Once()
		f.productRepo.EXPECT().GetBySlug(ctx, "cotton-shirt-2").Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.productRepo.EXPECT().Create(ctx, mock.MatchedBy(func(p *models.Product) bool {
			return p.Name == "Cotton Shirt" && p.Slug == "cotton-shirt-2" && p.CategoryID == 1 && p.BrandID == 2
		})).Return(nil).Once()

		got, err := f.uc.CreateProduct(ctx, "  Cotton Shirt ", &shortDesc, &desc, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, "cotton-shirt-2", got.Slug)
	})

	t.Run("create repo error", func(t *testing.T) {
		f := newProductFixture(t)
		dbErr := errors.New("insert failed")
		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Category{ID: 1, IsActive: true}, nil).Once()
		f.brandRepo.EXPECT().GetByID(ctx, int64(2)).Return(&models.Brand{ID: 2, IsActive: true}, nil).Once()
		f.productRepo.EXPECT().GetBySlug(ctx, "cotton-shirt").Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.productRepo.EXPECT().Create(ctx, mock.AnythingOfType("*models.Product")).Return(dbErr).Once()

		got, err := f.uc.CreateProduct(ctx, "Cotton Shirt", nil, nil, 1, 2)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUseCase_GetAndUpdateAndState(t *testing.T) {
	ctx := context.Background()

	t.Run("GetProduct and GetProductBySlug", func(t *testing.T) {
		f := newProductFixture(t)
		_, err := f.uc.GetProduct(ctx, 0)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductID)

		expected := &models.Product{ID: 5, Slug: "shirt"}
		f.productRepo.EXPECT().GetByID(ctx, uint(5)).Return(expected, nil).Once()
		got, err := f.uc.GetProduct(ctx, 5)
		require.NoError(t, err)
		assert.Equal(t, expected, got)

		_, err = f.uc.GetProductBySlug(ctx, "   ")
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductSlug)

		f.productRepo.EXPECT().GetBySlug(ctx, "shirt").Return(expected, nil).Once()
		got, err = f.uc.GetProductBySlug(ctx, " shirt ")
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("UpdateProduct", func(t *testing.T) {
		f := newProductFixture(t)
		_, err := f.uc.UpdateProduct(ctx, 0, "Valid Name", nil, nil, 1, 1)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductID)

		_, err = f.uc.UpdateProduct(ctx, 5, "", nil, nil, 1, 1)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductName)

		existing := &models.Product{ID: 5, Name: "Old", Slug: "old", CategoryID: 1, BrandID: 1}
		f.productRepo.EXPECT().GetByID(ctx, uint(5)).Return(existing, nil).Once()
		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Category{ID: 1, IsActive: true}, nil).Once()
		f.brandRepo.EXPECT().GetByID(ctx, int64(2)).Return(&models.Brand{ID: 2, IsActive: true}, nil).Once()
		f.productRepo.EXPECT().GetBySlug(ctx, "updated-shirt").Return(existing, nil).Once()
		f.productRepo.EXPECT().Update(ctx, mock.MatchedBy(func(p *models.Product) bool {
			return p.ID == 5 && p.Name == "Updated Shirt" && p.Slug == "updated-shirt" && p.BrandID == 2
		})).Return(nil).Once()

		got, err := f.uc.UpdateProduct(ctx, 5, "Updated Shirt", nil, nil, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, "updated-shirt", got.Slug)
	})

	t.Run("Delete Restore ToggleActive ToggleListed", func(t *testing.T) {
		f := newProductFixture(t)
		assert.ErrorIs(t, f.uc.DeleteProduct(ctx, 0), domainErrors.ErrInvalidProductID)
		assert.ErrorIs(t, f.uc.RestoreProduct(ctx, 0), domainErrors.ErrInvalidProductID)
		assert.ErrorIs(t, f.uc.ToggleProductActive(ctx, 0), domainErrors.ErrInvalidProductID)
		assert.ErrorIs(t, f.uc.ToggleProductListed(ctx, 0), domainErrors.ErrInvalidProductID)

		f.productRepo.EXPECT().GetByID(ctx, uint(7)).Return(&models.Product{ID: 7}, nil).Once()
		f.productRepo.EXPECT().SoftDelete(ctx, uint(7)).Return(nil).Once()
		require.NoError(t, f.uc.DeleteProduct(ctx, 7))

		f.productRepo.EXPECT().GetDeletedByID(ctx, uint(7)).Return(&models.Product{ID: 7, CategoryID: 1, BrandID: 2}, nil).Once()
		f.categoryRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.Category{ID: 1, IsActive: true}, nil).Once()
		f.brandRepo.EXPECT().GetByID(ctx, int64(2)).Return(&models.Brand{ID: 2, IsActive: true}, nil).Once()
		f.productRepo.EXPECT().Restore(ctx, uint(7)).Return(nil).Once()
		require.NoError(t, f.uc.RestoreProduct(ctx, 7))

		f.productRepo.EXPECT().GetByID(ctx, uint(7)).Return(&models.Product{ID: 7}, nil).Once()
		f.productRepo.EXPECT().ToggleActive(ctx, uint(7)).Return(nil).Once()
		require.NoError(t, f.uc.ToggleProductActive(ctx, 7))

		f.productRepo.EXPECT().GetByID(ctx, uint(7)).Return(&models.Product{ID: 7}, nil).Once()
		f.productRepo.EXPECT().ToggleListed(ctx, uint(7)).Return(nil).Once()
		require.NoError(t, f.uc.ToggleProductListed(ctx, 7))
	})
}

func TestProductUseCase_ListsAndStorefront(t *testing.T) {
	ctx := context.Background()

	t.Run("ListProducts and ListDeletedProducts", func(t *testing.T) {
		f := newProductFixture(t)
		items := []models.Product{{ID: 1}}
		// Note: ListProducts calls repo.List with (ctx, search, page, limit, ...)
		f.productRepo.EXPECT().List(ctx, "shoe", 1, 10, (*bool)(nil), (*bool)(nil), (*int64)(nil), (*int64)(nil)).
			Return(items, int64(1), nil).Once()

		got, total, err := f.uc.ListProducts(ctx, " shoe ", 0, 0, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, items, got)
		assert.Equal(t, int64(1), total)

		f.productRepo.EXPECT().ListDeleted(ctx, "old", 1, 10).Return(items, int64(1), nil).Once()
		gotDel, totDel, err := f.uc.ListDeletedProducts(ctx, " old ", 0, 0)
		require.NoError(t, err)
		assert.Equal(t, items, gotDel)
		assert.Equal(t, int64(1), totDel)
	})

	t.Run("ListStoreProducts validation and success", func(t *testing.T) {
		f := newProductFixture(t)
		badCat := int64(0)
		_, _, err := f.uc.ListStoreProducts(ctx, "", &badCat, nil, nil, nil, "", 1, 12)
		assert.Error(t, err)

		badBrand := int64(-1)
		_, _, err = f.uc.ListStoreProducts(ctx, "", nil, &badBrand, nil, nil, "", 1, 12)
		assert.Error(t, err)

		negPrice := -5.0
		_, _, err = f.uc.ListStoreProducts(ctx, "", nil, nil, &negPrice, nil, "", 1, 12)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidPriceRange)

		minP, maxP := 500.0, 100.0
		_, _, err = f.uc.ListStoreProducts(ctx, "", nil, nil, &minP, &maxP, "", 1, 12)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidPriceRange)

		_, _, err = f.uc.ListStoreProducts(ctx, "", nil, nil, nil, nil, "invalid_sort", 1, 12)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductSort)

		expected := []models.Product{{ID: 3}}
		catID := int64(2)
		minValid, maxValid := 100.0, 500.0
		f.productRepo.EXPECT().ListStoreProducts(ctx, "tee", &catID, (*int64)(nil), &minValid, &maxValid, "price_asc", 1, 100).
			Return(expected, int64(1), nil).Once()

		got, total, err := f.uc.ListStoreProducts(ctx, " tee ", &catID, nil, &minValid, &maxValid, "price_asc", 0, 500)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
		assert.Equal(t, int64(1), total)
	})

	t.Run("GetStoreProductByID GetStoreProductBySlug and GetRelatedProducts", func(t *testing.T) {
		f := newProductFixture(t)
		_, err := f.uc.GetStoreProductByID(ctx, 0)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductID)

		prod := &models.Product{ID: 4}
		f.productRepo.EXPECT().GetStoreProductByID(ctx, int64(4)).Return(prod, nil).Once()
		got, err := f.uc.GetStoreProductByID(ctx, 4)
		require.NoError(t, err)
		assert.Equal(t, prod, got)

		_, err = f.uc.GetStoreProductBySlug(ctx, "")
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductSlug)

		f.productRepo.EXPECT().GetStoreProductBySlug(ctx, "polo").Return(prod, nil).Once()
		got, err = f.uc.GetStoreProductBySlug(ctx, "polo")
		require.NoError(t, err)
		assert.Equal(t, prod, got)

		_, err = f.uc.GetRelatedProducts(ctx, 0, 5)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProductID)

		related := []models.Product{{ID: 8}}
		f.productRepo.EXPECT().GetRelatedProducts(ctx, int64(4), 8).Return(related, nil).Once()
		gotRel, err := f.uc.GetRelatedProducts(ctx, 4, 0)
		require.NoError(t, err)
		assert.Equal(t, related, gotRel)

		f.productRepo.EXPECT().GetRelatedProducts(ctx, int64(4), 20).Return(related, nil).Once()
		gotRel, err = f.uc.GetRelatedProducts(ctx, 4, 99)
		require.NoError(t, err)
		assert.Equal(t, related, gotRel)
	})
}
