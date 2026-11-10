package report

import (
	"errors"
	"testing"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
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
