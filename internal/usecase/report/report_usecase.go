package report

import (
	"context"
	"log/slog"
	"strings"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
)

type ReportUseCase struct {
	repo interfaces.ReportRepository
	log  *slog.Logger
}

func NewReportUseCase(
	repo interfaces.ReportRepository,
	log *slog.Logger,
) *ReportUseCase {
	return &ReportUseCase{
		repo: repo,
		log:  log,
	}
}
func validateFilter(filter *models.ReportFilter) error {
	filter.GroupBy = strings.ToLower(strings.TrimSpace(filter.GroupBy))
	if filter.GroupBy == "" {
		filter.GroupBy = "day"
	}
	switch filter.GroupBy {
	case "day", "week", "month", "year":
	default:
		return domainerrors.ErrInvalidInput
	}
	if !filter.StartDate.IsZero() &&
		!filter.EndDate.IsZero() &&
		filter.StartDate.After(filter.EndDate) {
		return domainerrors.ErrInvalidInput
	}
	return nil
}
func (u *ReportUseCase) GetSalesReport(ctx context.Context, filter models.ReportFilter) (*models.SalesReport, error) {
	if err := validateFilter(&filter); err != nil {
		return nil, err
	}
	result, err := u.repo.GetSalesReport(ctx, filter)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to generate sales report",
			"error", err,
		)
		return nil, err
	}
	return result, nil
}
func (u *ReportUseCase) GetOrdersReport(ctx context.Context, filter models.ReportFilter) (*models.OrdersReport, error) {
	if err := validateFilter(&filter); err != nil {
		return nil, err
	}
	result, err := u.repo.GetOrdersReport(ctx, filter)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to generate orders report",
			"error", err,
		)
		return nil, err
	}

	return result, nil
}
func (u *ReportUseCase) GetProductsReport(ctx context.Context, filter models.ReportFilter) (*models.ProductsReport, error) {
	if !filter.StartDate.IsZero() &&
		!filter.EndDate.IsZero() &&
		filter.StartDate.After(filter.EndDate) {
		return nil, domainerrors.ErrInvalidInput
	}
	result, err := u.repo.GetProductsReport(ctx, filter)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to generate products report",
			"error", err,
		)
		return nil, err
	}
	return result, nil
}

var _ interfaces.ReportUseCase = (*ReportUseCase)(nil)
