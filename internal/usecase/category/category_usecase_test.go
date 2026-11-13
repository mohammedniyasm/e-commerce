package category

import (
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
	uc           *CategoryUseCase
	categoryRepo *mocks.MockCategoryRepository
	productRepo  *mocks.MockProductRepository
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	catRepo := mocks.NewMockCategoryRepository(t)
	prodRepo := mocks.NewMockProductRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &fixture{
		uc:           NewCategoryUseCase(catRepo, prodRepo, log),
		categoryRepo: catRepo,
		productRepo:  prodRepo,
	}
}

func TestCategoryUseCase_CreateCategory(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		f := newFixture(t)
		got, err := f.uc.CreateCategory(context.Background(), "  ", "desc")
		require.ErrorIs(t, err, domainerrors.ErrInvalidCategoryName)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("insert failed")
		f.categoryRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.Category")).Return(dbErr).Once()

		got, err := f.uc.CreateCategory(context.Background(), "Shoes", "Footwear")
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *models.Category) bool {
			return c.Name == "Shoes" && c.Description != nil && *c.Description == "Footwear" && c.IsActive
		})).Run(func(_ context.Context, c *models.Category) {
			c.ID = 15
		}).Return(nil).Once()

		got, err := f.uc.CreateCategory(context.Background(), " Shoes ", " Footwear ")
		require.NoError(t, err)
		assert.Equal(t, int64(15), got.ID)
	})
}

func TestCategoryUseCase_GetCategory(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		got, err := f.uc.GetCategory(context.Background(), 0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidCategoryID)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(4)).Return(nil, domainerrors.ErrRecordNotFound).Once()

		got, err := f.uc.GetCategory(context.Background(), 4)
		require.ErrorIs(t, err, domainerrors.ErrRecordNotFound)
		assert.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		want := &models.Category{ID: 4, Name: "Shoes"}
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(4)).Return(want, nil).Once()

		got, err := f.uc.GetCategory(context.Background(), 4)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestCategoryUseCase_UpdateCategory(t *testing.T) {
	t.Run("validation errors", func(t *testing.T) {
		f := newFixture(t)
		_, err := f.uc.UpdateCategory(context.Background(), 0, "Shoes", "")
		require.ErrorIs(t, err, domainerrors.ErrInvalidCategoryID)

		_, err = f.uc.UpdateCategory(context.Background(), 1, "1", "")
		require.ErrorIs(t, err, domainerrors.ErrInvalidCategoryName)
	})

	t.Run("not found", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, domainerrors.ErrRecordNotFound).Once()

		_, err := f.uc.UpdateCategory(context.Background(), 1, "Shoes", "")
		require.ErrorIs(t, err, domainerrors.ErrRecordNotFound)
	})

	t.Run("duplicate name on different category", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&models.Category{ID: 1, Name: "Old"}, nil).Once()
		f.categoryRepo.EXPECT().GetByExactName(mock.Anything, "Shoes").Return(&models.Category{ID: 2, Name: "Shoes"}, nil).Once()

		_, err := f.uc.UpdateCategory(context.Background(), 1, "Shoes", "")
		require.ErrorIs(t, err, domainerrors.ErrCategoryNameAlreadyExists)
	})

	t.Run("update error and success", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("update failed")
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&models.Category{ID: 1, Name: "Old"}, nil).Twice()
		f.categoryRepo.EXPECT().GetByExactName(mock.Anything, "Shoes").Return(nil, domainerrors.ErrRecordNotFound).Twice()

		f.categoryRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*models.Category")).Return(dbErr).Once()
		_, err := f.uc.UpdateCategory(context.Background(), 1, "Shoes", "Updated")
		require.ErrorIs(t, err, dbErr)

		f.categoryRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *models.Category) bool {
			return c.ID == 1 && c.Name == "Shoes" && c.Description != nil && *c.Description == "Updated"
		})).Return(nil).Once()
		got, err := f.uc.UpdateCategory(context.Background(), 1, " Shoes ", " Updated ")
		require.NoError(t, err)
		assert.Equal(t, "Shoes", got.Name)
	})
}

func TestCategoryUseCase_DeleteCategory(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.DeleteCategory(context.Background(), -1), domainerrors.ErrInvalidCategoryID)
	})

	t.Run("category not found", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(3)).Return(nil, domainerrors.ErrRecordNotFound).Once()
		require.ErrorIs(t, f.uc.DeleteCategory(context.Background(), 3), domainerrors.ErrRecordNotFound)
	})

	t.Run("count active products fails", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("count error")
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(3)).Return(&models.Category{ID: 3}, nil).Once()
		f.productRepo.EXPECT().CountActiveByCategoryID(mock.Anything, int64(3)).Return(int64(0), dbErr).Once()
		require.ErrorIs(t, f.uc.DeleteCategory(context.Background(), 3), dbErr)
	})

	t.Run("category has active products", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(3)).Return(&models.Category{ID: 3}, nil).Once()
		f.productRepo.EXPECT().CountActiveByCategoryID(mock.Anything, int64(3)).Return(int64(2), nil).Once()
		require.ErrorIs(t, f.uc.DeleteCategory(context.Background(), 3), domainerrors.ErrCategoryHasActiveProducts)
	})

	t.Run("soft delete error and success", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("soft delete failed")
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(3)).Return(&models.Category{ID: 3}, nil).Twice()
		f.productRepo.EXPECT().CountActiveByCategoryID(mock.Anything, int64(3)).Return(int64(0), nil).Twice()

		f.categoryRepo.EXPECT().SoftDelete(mock.Anything, int64(3)).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.DeleteCategory(context.Background(), 3), dbErr)

		f.categoryRepo.EXPECT().SoftDelete(mock.Anything, int64(3)).Return(nil).Once()
		require.NoError(t, f.uc.DeleteCategory(context.Background(), 3))
	})
}

func TestCategoryUseCase_ListAndListDeleted(t *testing.T) {
	t.Run("list categories", func(t *testing.T) {
		f := newFixture(t)
		want := []models.Category{{ID: 1, Name: "Shoes"}}
		f.categoryRepo.EXPECT().List(mock.Anything, "shoe", 1, 100).Return(want, int64(1), nil).Once()

		got, total, err := f.uc.ListCategories(context.Background(), " shoe ", 0, 200)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, want, got)

		dbErr := errors.New("list failed")
		f.categoryRepo.EXPECT().List(mock.Anything, "", 1, 10).Return(nil, int64(0), dbErr).Once()
		_, _, err = f.uc.ListCategories(context.Background(), "", 0, 0)
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("list deleted categories", func(t *testing.T) {
		f := newFixture(t)
		want := []models.Category{{ID: 2, Name: "Old"}}
		f.categoryRepo.EXPECT().ListDeleted(mock.Anything, "old", 1, 10).Return(want, int64(1), nil).Once()

		got, total, err := f.uc.ListDeletedCategories(context.Background(), " old ", -1, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, want, got)

		dbErr := errors.New("list deleted failed")
		f.categoryRepo.EXPECT().ListDeleted(mock.Anything, "", 2, 100).Return(nil, int64(0), dbErr).Once()
		_, _, err = f.uc.ListDeletedCategories(context.Background(), "", 2, 500)
		require.ErrorIs(t, err, dbErr)
	})
}

func TestCategoryUseCase_RestoreCategory(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.RestoreCategory(context.Background(), 0), domainerrors.ErrInvalidCategoryID)
	})

	t.Run("deleted category not found", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().GetDeletedByID(mock.Anything, int64(5)).Return(nil, domainerrors.ErrRecordNotFound).Once()
		require.ErrorIs(t, f.uc.RestoreCategory(context.Background(), 5), domainerrors.ErrRecordNotFound)
	})

	t.Run("duplicate active category exists", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().GetDeletedByID(mock.Anything, int64(5)).Return(&models.Category{ID: 5, Name: "Shoes"}, nil).Once()
		f.categoryRepo.EXPECT().GetByExactName(mock.Anything, "Shoes").Return(&models.Category{ID: 9, Name: "Shoes"}, nil).Once()
		require.ErrorIs(t, f.uc.RestoreCategory(context.Background(), 5), domainerrors.ErrRestoreCategoryAlreadyExists)
	})

	t.Run("restore error and success", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("restore failed")
		f.categoryRepo.EXPECT().GetDeletedByID(mock.Anything, int64(5)).Return(&models.Category{ID: 5, Name: "Shoes"}, nil).Twice()
		f.categoryRepo.EXPECT().GetByExactName(mock.Anything, "Shoes").Return(nil, domainerrors.ErrRecordNotFound).Twice()

		f.categoryRepo.EXPECT().Restore(mock.Anything, int64(5)).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.RestoreCategory(context.Background(), 5), dbErr)

		f.categoryRepo.EXPECT().Restore(mock.Anything, int64(5)).Return(nil).Once()
		require.NoError(t, f.uc.RestoreCategory(context.Background(), 5))
	})
}

func TestCategoryUseCase_ToggleCategoryActive(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.ToggleCategoryActive(context.Background(), 0), domainerrors.ErrInvalidCategoryID)
	})

	t.Run("not found", func(t *testing.T) {
		f := newFixture(t)
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(2)).Return(nil, domainerrors.ErrRecordNotFound).Once()
		require.ErrorIs(t, f.uc.ToggleCategoryActive(context.Background(), 2), domainerrors.ErrRecordNotFound)
	})

	t.Run("toggle error and success", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("toggle failed")
		f.categoryRepo.EXPECT().GetByID(mock.Anything, int64(2)).Return(&models.Category{ID: 2, IsActive: true}, nil).Twice()
		f.categoryRepo.EXPECT().ToggleActive(mock.Anything, int64(2)).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.ToggleCategoryActive(context.Background(), 2), dbErr)

		f.categoryRepo.EXPECT().ToggleActive(mock.Anything, int64(2)).Return(nil).Once()
		require.NoError(t, f.uc.ToggleCategoryActive(context.Background(), 2))
	})
}

func TestCategoryUseCase_ListStoreCategories(t *testing.T) {
	t.Run("repo error wrapped", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("query failed")
		f.categoryRepo.EXPECT().ListStoreCategories(mock.Anything).Return(nil, dbErr).Once()

		got, err := f.uc.ListStoreCategories(context.Background())
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		want := []models.Category{{ID: 1, Name: "Shoes"}}
		f.categoryRepo.EXPECT().ListStoreCategories(mock.Anything).Return(want, nil).Once()

		got, err := f.uc.ListStoreCategories(context.Background())
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
