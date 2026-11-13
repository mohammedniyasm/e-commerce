package report

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    *models.ReportFilter
		wantGroup string
		wantError error
	}{
		// Default value
		{
			name: "empty group by defaults to day",
			filter: &models.ReportFilter{
				GroupBy: "",
			},
			wantGroup: "day",
			wantError: nil,
		},

		// Valid groupings
		{
			name: "valid day grouping",
			filter: &models.ReportFilter{
				GroupBy: "day",
			},
			wantGroup: "day",
			wantError: nil,
		},
		{
			name: "valid week grouping",
			filter: &models.ReportFilter{
				GroupBy: "week",
			},
			wantGroup: "week",
			wantError: nil,
		},
		{
			name: "valid month grouping",
			filter: &models.ReportFilter{
				GroupBy: "month",
			},
			wantGroup: "month",
			wantError: nil,
		},
		{
			name: "valid year grouping",
			filter: &models.ReportFilter{
				GroupBy: "year",
			},
			wantGroup: "year",
			wantError: nil,
		},

		// Case normalization
		{
			name: "uppercase group by",
			filter: &models.ReportFilter{
				GroupBy: "MONTH",
			},
			wantGroup: "month",
			wantError: nil,
		},
		{
			name: "mixed case group by",
			filter: &models.ReportFilter{
				GroupBy: "WeEk",
			},
			wantGroup: "week",
			wantError: nil,
		},

		// Whitespace handling
		{
			name: "group by with surrounding spaces",
			filter: &models.ReportFilter{
				GroupBy: "  month  ",
			},
			wantGroup: "month",
			wantError: nil,
		},
		{
			name: "group by containing only spaces",
			filter: &models.ReportFilter{
				GroupBy: "   ",
			},
			wantGroup: "day",
			wantError: nil,
		},

		// Invalid groupings
		{
			name: "invalid group by",
			filter: &models.ReportFilter{
				GroupBy: "hour",
			},
			wantGroup: "hour",
			wantError: domainerrors.ErrInvalidInput,
		},

		// Date validation
		{
			name: "start date before end date",
			filter: &models.ReportFilter{
				GroupBy:   "day",
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			},
			wantGroup: "day",
			wantError: nil,
		},
		{
			name: "start date equals end date",
			filter: &models.ReportFilter{
				GroupBy:   "day",
				StartDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			},
			wantGroup: "day",
			wantError: nil,
		},
		{
			name: "start date after end date",
			filter: &models.ReportFilter{
				GroupBy:   "day",
				StartDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			},
			wantGroup: "day",
			wantError: domainerrors.ErrInvalidInput,
		},
		{
			name: "only start date provided",
			filter: &models.ReportFilter{
				GroupBy:   "day",
				StartDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			},
			wantGroup: "day",
			wantError: nil,
		},
		{
			name: "only end date provided",
			filter: &models.ReportFilter{
				GroupBy: "day",
				EndDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			},
			wantGroup: "day",
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilter(tt.filter)

			if !errors.Is(err, tt.wantError) {
				t.Errorf(
					"validateFilter() error = %v; want %v",
					err,
					tt.wantError,
				)
			}

			if tt.filter.GroupBy != tt.wantGroup {
				t.Errorf(
					"GroupBy = %q; want %q",
					tt.filter.GroupBy,
					tt.wantGroup,
				)
			}
		})
	}
}

type reportFixture struct {
	uc   *ReportUseCase
	repo *mocks.MockReportRepository
}

func newReportFixture(t *testing.T) *reportFixture {
	t.Helper()
	repo := mocks.NewMockReportRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &reportFixture{
		uc:   NewReportUseCase(repo, log),
		repo: repo,
	}
}

func TestReportUseCase_GetSalesReport(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("invalid filter returns error without calling repo", func(t *testing.T) {
		f := newReportFixture(t)
		got, err := f.uc.GetSalesReport(ctx, models.ReportFilter{GroupBy: "hourly"})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		got, err = f.uc.GetSalesReport(ctx, models.ReportFilter{
			StartDate: now.Add(time.Hour),
			EndDate:   now,
		})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	t.Run("normalizes empty GroupBy to day and returns report", func(t *testing.T) {
		f := newReportFixture(t)
		expected := &models.SalesReport{}
		f.repo.EXPECT().GetSalesReport(ctx, models.ReportFilter{GroupBy: "day"}).Return(expected, nil).Once()

		got, err := f.uc.GetSalesReport(ctx, models.ReportFilter{})
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newReportFixture(t)
		dbErr := errors.New("sales report failed")
		f.repo.EXPECT().GetSalesReport(ctx, models.ReportFilter{GroupBy: "month"}).Return(nil, dbErr).Once()

		got, err := f.uc.GetSalesReport(ctx, models.ReportFilter{GroupBy: "month"})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestReportUseCase_GetOrdersReport(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid group_by returns invalid input", func(t *testing.T) {
		f := newReportFixture(t)
		got, err := f.uc.GetOrdersReport(ctx, models.ReportFilter{GroupBy: "invalid"})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	t.Run("success and repo error", func(t *testing.T) {
		f := newReportFixture(t)
		expected := &models.OrdersReport{}
		f.repo.EXPECT().GetOrdersReport(ctx, models.ReportFilter{GroupBy: "day"}).Return(expected, nil).Once()

		got, err := f.uc.GetOrdersReport(ctx, models.ReportFilter{})
		require.NoError(t, err)
		assert.Equal(t, expected, got)

		dbErr := errors.New("orders report failed")
		f.repo.EXPECT().GetOrdersReport(ctx, models.ReportFilter{GroupBy: "week"}).Return(nil, dbErr).Once()

		got, err = f.uc.GetOrdersReport(ctx, models.ReportFilter{GroupBy: "week"})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestReportUseCase_GetProductsReport(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("start date after end date returns invalid input", func(t *testing.T) {
		f := newReportFixture(t)
		got, err := f.uc.GetProductsReport(ctx, models.ReportFilter{
			StartDate: now.Add(time.Hour),
			EndDate:   now,
		})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	t.Run("skips GroupBy validation and delegates directly", func(t *testing.T) {
		f := newReportFixture(t)
		filter := models.ReportFilter{GroupBy: "custom", StartDate: now.Add(-time.Hour), EndDate: now}
		expected := &models.ProductsReport{}
		f.repo.EXPECT().GetProductsReport(ctx, filter).Return(expected, nil).Once()

		got, err := f.uc.GetProductsReport(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, expected, got)

		dbErr := errors.New("products report failed")
		f.repo.EXPECT().GetProductsReport(ctx, filter).Return(nil, dbErr).Once()
		got, err = f.uc.GetProductsReport(ctx, filter)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}
