package dashboard

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	domainErrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dashboardFixture struct {
	uc   *DashboardUseCase
	repo *mocks.MockDashboardRepository
}

func newDashboardFixture(t *testing.T) *dashboardFixture {
	t.Helper()
	repo := mocks.NewMockDashboardRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &dashboardFixture{
		uc:   NewDashboardUseCase(repo, log),
		repo: repo,
	}
}

func TestDashboardUseCase_GetOverview(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		f := newDashboardFixture(t)
		expected := &models.DashboardOverview{}
		f.repo.EXPECT().GetOverview(ctx).Return(expected, nil).Once()

		got, err := f.uc.GetOverview(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newDashboardFixture(t)
		dbErr := errors.New("db error")
		f.repo.EXPECT().GetOverview(ctx).Return(nil, dbErr).Once()

		got, err := f.uc.GetOverview(ctx)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestDashboardUseCase_GetSales(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name   string
			filter models.DashboardFilter
		}{
			{
				name:   "invalid group_by",
				filter: models.DashboardFilter{GroupBy: "hour"},
			},
			{
				name: "start date after end date",
				filter: models.DashboardFilter{
					StartDate: now.Add(24 * time.Hour),
					EndDate:   now,
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				f := newDashboardFixture(t)
				got, err := f.uc.GetSales(ctx, tc.filter)
				assert.Nil(t, got)
				assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
			})
		}
	})

	t.Run("normalizes group_by and delegates", func(t *testing.T) {
		tests := []struct {
			name          string
			inputGroupBy  string
			expectGroupBy string
		}{
			{name: "empty defaults to day", inputGroupBy: "", expectGroupBy: "day"},
			{name: "whitespace defaults to day", inputGroupBy: "   ", expectGroupBy: "day"},
			{name: "trims and lowercases week", inputGroupBy: "  WEEK ", expectGroupBy: "week"},
			{name: "month", inputGroupBy: "month", expectGroupBy: "month"},
			{name: "year", inputGroupBy: "Year", expectGroupBy: "year"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				f := newDashboardFixture(t)
				input := models.DashboardFilter{
					GroupBy:   tc.inputGroupBy,
					StartDate: now.Add(-48 * time.Hour),
					EndDate:   now,
				}
				expectedFilter := input
				expectedFilter.GroupBy = tc.expectGroupBy
				expectedSales := []models.SalesData{{}}

				f.repo.EXPECT().GetSales(ctx, expectedFilter).Return(expectedSales, nil).Once()

				got, err := f.uc.GetSales(ctx, input)
				require.NoError(t, err)
				assert.Equal(t, expectedSales, got)
			})
		}
	})

	t.Run("only start date provided skips range ordering error", func(t *testing.T) {
		f := newDashboardFixture(t)
		input := models.DashboardFilter{StartDate: now}
		expectedFilter := models.DashboardFilter{StartDate: now, GroupBy: "day"}
		f.repo.EXPECT().GetSales(ctx, expectedFilter).Return([]models.SalesData{}, nil).Once()

		got, err := f.uc.GetSales(ctx, input)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newDashboardFixture(t)
		dbErr := errors.New("query failed")
		f.repo.EXPECT().GetSales(ctx, models.DashboardFilter{GroupBy: "day"}).Return(nil, dbErr).Once()

		got, err := f.uc.GetSales(ctx, models.DashboardFilter{})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestDashboardUseCase_TopLists(t *testing.T) {
	ctx := context.Background()

	limits := []struct {
		name        string
		input       int
		expectLimit int
	}{
		{name: "zero defaults to 10", input: 0, expectLimit: 10},
		{name: "negative defaults to 10", input: -3, expectLimit: 10},
		{name: "above 100 capped to 100", input: 250, expectLimit: 100},
		{name: "valid kept as is", input: 25, expectLimit: 25},
	}

	t.Run("GetTopProducts", func(t *testing.T) {
		for _, tc := range limits {
			t.Run(tc.name, func(t *testing.T) {
				f := newDashboardFixture(t)
				expected := []models.TopProduct{{}}
				f.repo.EXPECT().GetTopProducts(ctx, tc.expectLimit).Return(expected, nil).Once()

				got, err := f.uc.GetTopProducts(ctx, tc.input)
				require.NoError(t, err)
				assert.Equal(t, expected, got)
			})
		}

		t.Run("repo error", func(t *testing.T) {
			f := newDashboardFixture(t)
			dbErr := errors.New("db error")
			f.repo.EXPECT().GetTopProducts(ctx, 10).Return(nil, dbErr).Once()

			got, err := f.uc.GetTopProducts(ctx, 0)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, dbErr)
		})
	})

	t.Run("GetTopCategories", func(t *testing.T) {
		for _, tc := range limits {
			t.Run(tc.name, func(t *testing.T) {
				f := newDashboardFixture(t)
				expected := []models.TopCategory{{}}
				f.repo.EXPECT().GetTopCategories(ctx, tc.expectLimit).Return(expected, nil).Once()

				got, err := f.uc.GetTopCategories(ctx, tc.input)
				require.NoError(t, err)
				assert.Equal(t, expected, got)
			})
		}

		t.Run("repo error", func(t *testing.T) {
			f := newDashboardFixture(t)
			dbErr := errors.New("db error")
			f.repo.EXPECT().GetTopCategories(ctx, 10).Return(nil, dbErr).Once()

			got, err := f.uc.GetTopCategories(ctx, 0)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, dbErr)
		})
	})

	t.Run("GetTopBrands", func(t *testing.T) {
		for _, tc := range limits {
			t.Run(tc.name, func(t *testing.T) {
				f := newDashboardFixture(t)
				expected := []models.TopBrand{{}}
				f.repo.EXPECT().GetTopBrands(ctx, tc.expectLimit).Return(expected, nil).Once()

				got, err := f.uc.GetTopBrands(ctx, tc.input)
				require.NoError(t, err)
				assert.Equal(t, expected, got)
			})
		}

		t.Run("repo error", func(t *testing.T) {
			f := newDashboardFixture(t)
			dbErr := errors.New("db error")
			f.repo.EXPECT().GetTopBrands(ctx, 10).Return(nil, dbErr).Once()

			got, err := f.uc.GetTopBrands(ctx, 0)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, dbErr)
		})
	})
}

func TestDashboardUseCase_Stats(t *testing.T) {
	ctx := context.Background()

	t.Run("GetOrderStats success and error", func(t *testing.T) {
		f := newDashboardFixture(t)
		expected := &models.OrderStats{}
		f.repo.EXPECT().GetOrderStats(ctx).Return(expected, nil).Once()

		got, err := f.uc.GetOrderStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, got)

		dbErr := errors.New("stats failed")
		f.repo.EXPECT().GetOrderStats(ctx).Return(nil, dbErr).Once()

		got, err = f.uc.GetOrderStats(ctx)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("GetCustomerStats success and error", func(t *testing.T) {
		f := newDashboardFixture(t)
		expected := &models.CustomerStats{}
		f.repo.EXPECT().GetCustomerStats(ctx).Return(expected, nil).Once()

		got, err := f.uc.GetCustomerStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, got)

		dbErr := errors.New("stats failed")
		f.repo.EXPECT().GetCustomerStats(ctx).Return(nil, dbErr).Once()

		got, err = f.uc.GetCustomerStats(ctx)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}
