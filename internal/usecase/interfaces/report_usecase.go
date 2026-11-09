package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type ReportUseCase interface {
	GetSalesReport(ctx context.Context, filter models.ReportFilter) (*models.SalesReport, error)
	GetOrdersReport(ctx context.Context, filter models.ReportFilter) (*models.OrdersReport, error)
	GetProductsReport(ctx context.Context, filter models.ReportFilter) (*models.ProductsReport, error)
}
