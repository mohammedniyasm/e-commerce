package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type DashboardRepository interface {
	GetOverview(ctx context.Context) (*models.DashboardOverview, error)
	GetSales(ctx context.Context, filter models.DashboardFilter) ([]models.SalesData, error)
	GetOrderStats(ctx context.Context) (*models.OrderStats, error)
	GetCustomerStats(ctx context.Context) (*models.CustomerStats, error)
	GetTopProducts(ctx context.Context, limit int) ([]models.TopProduct, error)
	GetTopCategories(ctx context.Context, limit int) ([]models.TopCategory, error)
	GetTopBrands(ctx context.Context, limit int) ([]models.TopBrand, error)
}
