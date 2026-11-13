package brand

import (
	"bytes"
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	uc          *BrandUseCase
	brandRepo   *mocks.MockBrandRepository
	productRepo *mocks.MockProductRepository
	storage     *mocks.MockObjectStorage
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	brandRepo := mocks.NewMockBrandRepository(t)
	productRepo := mocks.NewMockProductRepository(t)
	storage := mocks.NewMockObjectStorage(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &fixture{
		uc:          NewBrandUseCase(brandRepo, productRepo, storage, log),
		brandRepo:   brandRepo,
		productRepo: productRepo,
		storage:     storage,
	}
}

func TestBrandUseCase_CreateBrand(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		f := newFixture(t)
		got, err := f.uc.CreateBrand(context.Background(), "   ", "desc")
		require.ErrorIs(t, err, domainerrors.ErrInvalidBrandName)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("insert failed")
		f.brandRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.Brand")).Return(dbErr).Once()

		got, err := f.uc.CreateBrand(context.Background(), "Nike", "Sportswear")
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(b *models.Brand) bool {
			return b.Name == "Nike" && b.Description != nil && *b.Description == "Sportswear" && b.IsActive
		})).Run(func(_ context.Context, b *models.Brand) {
			b.ID = 11
		}).Return(nil).Once()

		got, err := f.uc.CreateBrand(context.Background(), "  Nike  ", "  Sportswear ")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint(11), got.ID)
	})
}

func TestBrandUseCase_GetBrand(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		got, err := f.uc.GetBrand(context.Background(), 0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidBrandID)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(3)).Return(nil, domainerrors.ErrRecordNotFound).Once()

		got, err := f.uc.GetBrand(context.Background(), 3)
		require.ErrorIs(t, err, domainerrors.ErrRecordNotFound)
		assert.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		want := &models.Brand{ID: 3, Name: "Puma"}
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(3)).Return(want, nil).Once()

		got, err := f.uc.GetBrand(context.Background(), 3)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestBrandUseCase_UpdateBrand(t *testing.T) {
	t.Run("invalid id or name", func(t *testing.T) {
		f := newFixture(t)
		_, err := f.uc.UpdateBrand(context.Background(), 0, "Nike", "")
		require.ErrorIs(t, err, domainerrors.ErrInvalidBrandID)

		_, err = f.uc.UpdateBrand(context.Background(), 1, "A", "")
		require.ErrorIs(t, err, domainerrors.ErrInvalidBrandName)
	})

	t.Run("brand not found", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(5)).Return(nil, domainerrors.ErrRecordNotFound).Once()

		_, err := f.uc.UpdateBrand(context.Background(), 5, "Adidas", "desc")
		require.ErrorIs(t, err, domainerrors.ErrRecordNotFound)
	})

	t.Run("duplicate name on another brand", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(5)).Return(&models.Brand{ID: 5, Name: "Old"}, nil).Once()
		f.brandRepo.EXPECT().GetByExactName(mock.Anything, "Adidas").Return(&models.Brand{ID: 8, Name: "Adidas"}, nil).Once()

		_, err := f.uc.UpdateBrand(context.Background(), 5, "Adidas", "desc")
		require.ErrorIs(t, err, domainerrors.ErrBrandNameAlreadyExists)
	})

	t.Run("update error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("update failed")
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(5)).Return(&models.Brand{ID: 5, Name: "Old"}, nil).Once()
		f.brandRepo.EXPECT().GetByExactName(mock.Anything, "Adidas").Return(nil, domainerrors.ErrRecordNotFound).Once()
		f.brandRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*models.Brand")).Return(dbErr).Once()

		_, err := f.uc.UpdateBrand(context.Background(), 5, "Adidas", "desc")
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("success keeping own name", func(t *testing.T) {
		f := newFixture(t)
		existing := &models.Brand{ID: 5, Name: "Adidas"}
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(5)).Return(existing, nil).Once()
		f.brandRepo.EXPECT().GetByExactName(mock.Anything, "Adidas").Return(existing, nil).Once()
		f.brandRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(b *models.Brand) bool {
			return b.ID == 5 && b.Name == "Adidas" && b.Description != nil && *b.Description == "Updated"
		})).Return(nil).Once()

		got, err := f.uc.UpdateBrand(context.Background(), 5, " Adidas ", " Updated ")
		require.NoError(t, err)
		assert.Equal(t, "Adidas", got.Name)
	})
}

func TestBrandUseCase_DeleteBrand(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.DeleteBrand(context.Background(), 0), domainerrors.ErrInvalidBrandID)
	})

	t.Run("get brand fails", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(2)).Return(nil, domainerrors.ErrRecordNotFound).Once()
		require.ErrorIs(t, f.uc.DeleteBrand(context.Background(), 2), domainerrors.ErrRecordNotFound)
	})

	t.Run("count active products fails", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("count error")
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(2)).Return(&models.Brand{ID: 2}, nil).Once()
		f.productRepo.EXPECT().CountActiveByBrandID(mock.Anything, int64(2)).Return(int64(0), dbErr).Once()
		require.ErrorIs(t, f.uc.DeleteBrand(context.Background(), 2), dbErr)
	})

	t.Run("brand has active products", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(2)).Return(&models.Brand{ID: 2}, nil).Once()
		f.productRepo.EXPECT().CountActiveByBrandID(mock.Anything, int64(2)).Return(int64(3), nil).Once()
		require.ErrorIs(t, f.uc.DeleteBrand(context.Background(), 2), domainerrors.ErrBrandHasActiveProducts)
	})

	t.Run("soft delete error and success", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("delete failed")
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(2)).Return(&models.Brand{ID: 2}, nil).Twice()
		f.productRepo.EXPECT().CountActiveByBrandID(mock.Anything, int64(2)).Return(int64(0), nil).Twice()
		f.brandRepo.EXPECT().SoftDelete(mock.Anything, int64(2)).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.DeleteBrand(context.Background(), 2), dbErr)

		f.brandRepo.EXPECT().SoftDelete(mock.Anything, int64(2)).Return(nil).Once()
		require.NoError(t, f.uc.DeleteBrand(context.Background(), 2))
	})
}

func TestBrandUseCase_ListBrandsAndDeletedBrands(t *testing.T) {
	t.Run("list brands normalizes args and handles errors", func(t *testing.T) {
		f := newFixture(t)
		want := []models.Brand{{ID: 1, Name: "Nike"}}
		f.brandRepo.EXPECT().List(mock.Anything, "nike", 1, 100).Return(want, int64(1), nil).Once()

		got, total, err := f.uc.ListBrands(context.Background(), "  nike ", 0, 500)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, want, got)

		dbErr := errors.New("list failed")
		f.brandRepo.EXPECT().List(mock.Anything, "", 1, 10).Return(nil, int64(0), dbErr).Once()
		_, _, err = f.uc.ListBrands(context.Background(), "", 0, 0)
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("list deleted brands normalizes args and handles errors", func(t *testing.T) {
		f := newFixture(t)
		want := []models.Brand{{ID: 2, Name: "Old"}}
		f.brandRepo.EXPECT().ListDeleted(mock.Anything, "old", 1, 10).Return(want, int64(1), nil).Once()

		got, total, err := f.uc.ListDeletedBrands(context.Background(), " old ", -1, -5)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, want, got)

		dbErr := errors.New("list deleted failed")
		f.brandRepo.EXPECT().ListDeleted(mock.Anything, "", 2, 100).Return(nil, int64(0), dbErr).Once()
		_, _, err = f.uc.ListDeletedBrands(context.Background(), "", 2, 150)
		require.ErrorIs(t, err, dbErr)
	})
}

func TestBrandUseCase_RestoreBrand(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.RestoreBrand(context.Background(), 0), domainerrors.ErrInvalidBrandID)
	})

	t.Run("deleted brand not found", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().GetDeletedByID(mock.Anything, int64(4)).Return(nil, domainerrors.ErrRecordNotFound).Once()
		require.ErrorIs(t, f.uc.RestoreBrand(context.Background(), 4), domainerrors.ErrRecordNotFound)
	})

	t.Run("active brand with same name exists", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().GetDeletedByID(mock.Anything, int64(4)).Return(&models.Brand{ID: 4, Name: "Nike"}, nil).Once()
		f.brandRepo.EXPECT().GetByExactName(mock.Anything, "Nike").Return(&models.Brand{ID: 9, Name: "Nike"}, nil).Once()
		require.ErrorIs(t, f.uc.RestoreBrand(context.Background(), 4), domainerrors.ErrRestoreBrandAlreadyExists)
	})

	t.Run("restore error and success", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("restore failed")
		f.brandRepo.EXPECT().GetDeletedByID(mock.Anything, int64(4)).Return(&models.Brand{ID: 4, Name: "Nike"}, nil).Twice()
		f.brandRepo.EXPECT().GetByExactName(mock.Anything, "Nike").Return(nil, domainerrors.ErrRecordNotFound).Twice()

		f.brandRepo.EXPECT().Restore(mock.Anything, int64(4)).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.RestoreBrand(context.Background(), 4), dbErr)

		f.brandRepo.EXPECT().Restore(mock.Anything, int64(4)).Return(nil).Once()
		require.NoError(t, f.uc.RestoreBrand(context.Background(), 4))
	})
}

func TestBrandUseCase_ToggleBrandActive(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.ToggleBrandActive(context.Background(), -1), domainerrors.ErrInvalidBrandID)
	})

	t.Run("not found", func(t *testing.T) {
		f := newFixture(t)
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, domainerrors.ErrRecordNotFound).Once()
		require.ErrorIs(t, f.uc.ToggleBrandActive(context.Background(), 1), domainerrors.ErrRecordNotFound)
	})

	t.Run("toggle error and success", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("toggle failed")
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&models.Brand{ID: 1, IsActive: true}, nil).Twice()
		f.brandRepo.EXPECT().ToggleActive(mock.Anything, int64(1)).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.ToggleBrandActive(context.Background(), 1), dbErr)

		f.brandRepo.EXPECT().ToggleActive(mock.Anything, int64(1)).Return(nil).Once()
		require.NoError(t, f.uc.ToggleBrandActive(context.Background(), 1))
	})
}

func TestBrandUseCase_LogoOperations(t *testing.T) {
	t.Run("upload logo validation and errors", func(t *testing.T) {
		f := newFixture(t)
		_, err := f.uc.UploadBrandLogo(context.Background(), 0, bytes.NewReader([]byte("x")), 1, "image/png")
		require.ErrorIs(t, err, domainerrors.ErrInvalidBrandID)

		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, domainerrors.ErrRecordNotFound).Once()
		_, err = f.uc.UploadBrandLogo(context.Background(), 1, bytes.NewReader([]byte("x")), 1, "image/png")
		require.ErrorIs(t, err, domainerrors.ErrRecordNotFound)

		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&models.Brand{ID: 1}, nil).Times(3)
		_, err = f.uc.UploadBrandLogo(context.Background(), 1, bytes.NewReader([]byte("x")), 10, "image/gif")
		require.ErrorIs(t, err, domainerrors.ErrInvalidBrandLogo)

		_, err = f.uc.UploadBrandLogo(context.Background(), 1, bytes.NewReader([]byte("x")), 0, "image/png")
		require.ErrorIs(t, err, domainerrors.ErrInvalidBrandLogoSize)

		_, err = f.uc.UploadBrandLogo(context.Background(), 1, bytes.NewReader([]byte("x")), 5*1024*1024+1, "image/webp")
		require.ErrorIs(t, err, domainerrors.ErrInvalidBrandLogoSize)
	})

	t.Run("upload logo storage and db failures", func(t *testing.T) {
		f := newFixture(t)
		upErr := errors.New("upload error")
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&models.Brand{ID: 1}, nil).Twice()
		f.storage.EXPECT().Upload(mock.Anything, "brands/1/logo", mock.Anything, int64(100), "image/jpeg").
			Return("", upErr).Once()
		_, err := f.uc.UploadBrandLogo(context.Background(), 1, bytes.NewReader(make([]byte, 100)), 100, "image/jpeg")
		require.ErrorIs(t, err, upErr)

		dbErr := errors.New("db error")
		logoURL := "https://cdn.example.com/brands/1/logo"
		f.storage.EXPECT().Upload(mock.Anything, "brands/1/logo", mock.Anything, int64(100), "image/png").
			Return(logoURL, nil).Once()
		f.brandRepo.EXPECT().UpdateLogo(mock.Anything, int64(1), &logoURL).Return(dbErr).Once()
		_, err = f.uc.UploadBrandLogo(context.Background(), 1, bytes.NewReader(make([]byte, 100)), 100, "image/png")
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("upload logo success", func(t *testing.T) {
		f := newFixture(t)
		logoURL := "https://cdn.example.com/brands/1/logo"
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&models.Brand{ID: 1}, nil).Once()
		f.storage.EXPECT().Upload(mock.Anything, "brands/1/logo", mock.Anything, int64(100), "image/webp").
			Return(logoURL, nil).Once()
		f.brandRepo.EXPECT().UpdateLogo(mock.Anything, int64(1), &logoURL).Return(nil).Once()

		got, err := f.uc.UploadBrandLogo(context.Background(), 1, bytes.NewReader(make([]byte, 100)), 100, "image/webp")
		require.NoError(t, err)
		require.NotNil(t, got.Logo)
		assert.Equal(t, logoURL, *got.Logo)
	})

	t.Run("delete logo cases", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.DeleteBrandLogo(context.Background(), 0), domainerrors.ErrInvalidBrandID)

		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, domainerrors.ErrRecordNotFound).Once()
		require.ErrorIs(t, f.uc.DeleteBrandLogo(context.Background(), 1), domainerrors.ErrRecordNotFound)

		// No existing logo -> noop
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&models.Brand{ID: 1, Logo: nil}, nil).Once()
		require.NoError(t, f.uc.DeleteBrandLogo(context.Background(), 1))

		// Storage delete error
		logo := "https://cdn.example.com/brands/1/logo"
		delErr := errors.New("s3 error")
		f.brandRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&models.Brand{ID: 1, Logo: &logo}, nil).Times(3)
		f.storage.EXPECT().Delete(mock.Anything, "brands/1/logo").Return(delErr).Once()
		require.ErrorIs(t, f.uc.DeleteBrandLogo(context.Background(), 1), delErr)

		// Repo UpdateLogo error
		dbErr := errors.New("db error")
		f.storage.EXPECT().Delete(mock.Anything, "brands/1/logo").Return(nil).Twice()
		f.brandRepo.EXPECT().UpdateLogo(mock.Anything, int64(1), (*string)(nil)).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.DeleteBrandLogo(context.Background(), 1), dbErr)

		// Success
		f.brandRepo.EXPECT().UpdateLogo(mock.Anything, int64(1), (*string)(nil)).Return(nil).Once()
		require.NoError(t, f.uc.DeleteBrandLogo(context.Background(), 1))
	})
}

func TestBrandUseCase_ListStoreBrands(t *testing.T) {
	t.Run("repo error wrapped", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("query failed")
		f.brandRepo.EXPECT().ListStoreBrands(mock.Anything).Return(nil, dbErr).Once()

		got, err := f.uc.ListStoreBrands(context.Background())
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		want := []models.Brand{{ID: 1, Name: "Nike"}}
		f.brandRepo.EXPECT().ListStoreBrands(mock.Anything).Return(want, nil).Once()

		got, err := f.uc.ListStoreBrands(context.Background())
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
