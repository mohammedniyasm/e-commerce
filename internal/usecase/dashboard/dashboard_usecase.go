package dashboard

import (
	"context"
	"log/slog"
	"strings"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
)

type DashboardUseCase struct {
	dashboardRepo interfaces.DashboardRepository
	log           *slog.Logger
}

func NewDashboardUseCase(
	dashboardRepo interfaces.DashboardRepository,
	log *slog.Logger,
) *DashboardUseCase {
	return &DashboardUseCase{
		dashboardRepo: dashboardRepo,
		log:           log,
	}
}
func (u *DashboardUseCase) GetOverview(ctx context.Context) (*models.DashboardOverview, error) {
	result, err := u.dashboardRepo.GetOverview(ctx)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to get dashboard overview",
			"error", err,
		)
		return nil, err
	}
	return result, nil
}
func (u *DashboardUseCase) GetSales(ctx context.Context, filter models.DashboardFilter) ([]models.SalesData, error) {
	filter.GroupBy = strings.ToLower(strings.TrimSpace(filter.GroupBy))
	if filter.GroupBy == "" {
		filter.GroupBy = "day"
	}
	switch filter.GroupBy {
	case "day", "week", "month", "year":
	default:
		return nil, domainerrors.ErrInvalidInput
	}
	if !filter.StartDate.IsZero() &&
		!filter.EndDate.IsZero() &&
		filter.StartDate.After(filter.EndDate) {
		return nil, domainerrors.ErrInvalidInput
	}
	result, err := u.dashboardRepo.GetSales(ctx, filter)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to get dashboard sales",
			"group_by", filter.GroupBy,
			"error", err,
		)
		return nil, err
	}
	return result, nil
}
func (u *DashboardUseCase) GetOrderStats(ctx context.Context) (*models.OrderStats, error) {
	result, err := u.dashboardRepo.GetOrderStats(ctx)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to get dashboard order statistics",
			"error", err,
		)
		return nil, err
	}
	return result, nil
}
func (u *DashboardUseCase) GetCustomerStats(ctx context.Context) (*models.CustomerStats, error) {
	result, err := u.dashboardRepo.GetCustomerStats(ctx)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to get dashboard customer statistics",
			"error", err,
		)
		return nil, err
	}
	return result, nil
}
func (u *DashboardUseCase) GetTopProducts(ctx context.Context, limit int) ([]models.TopProduct, error) {
	limit = normalizeLimit(limit)
	result, err := u.dashboardRepo.GetTopProducts(ctx, limit)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to get top products",
			"error", err,
		)
		return nil, err
	}
	return result, nil
}
func (u *DashboardUseCase) GetTopCategories(ctx context.Context, limit int) ([]models.TopCategory, error) {
	limit = normalizeLimit(limit)
	result, err := u.dashboardRepo.GetTopCategories(ctx, limit)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to get top categories",
			"error", err,
		)
		return nil, err
	}
	return result, nil
}
func (u *DashboardUseCase) GetTopBrands(ctx context.Context, limit int) ([]models.TopBrand, error) {
	limit = normalizeLimit(limit)
	result, err := u.dashboardRepo.GetTopBrands(ctx, limit)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to get top brands",
			"error", err,
		)
		return nil, err
	}
	return result, nil
}
func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 100 {
		return 100
	}
	return limit
}

var _ interfaces.DashboardUseCase = (*DashboardUseCase)(nil)
