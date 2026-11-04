package postgres

import (
	"context"
	"fmt"

	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{
		db: db,
	}
}

func (r *DashboardRepository) GetOverview(ctx context.Context) (*models.DashboardOverview, error) {
	var overview models.DashboardOverview
	query := `
	SELECT
		(
			SELECT COALESCE(SUM(total_amount), 0)
			FROM orders
			WHERE payment_status IN ('paid', 'partially_refunded')
		) AS total_revenue,
		COUNT(*) AS total_orders,
		(
			SELECT COUNT(*)
			FROM users
			WHERE role = 'user'
		) AS total_customers,
		(
			SELECT COUNT(*)
			FROM products
			WHERE deleted_at IS NULL
		) AS total_products,
		COUNT(*) FILTER (WHERE order_status = 'pending') AS pending_orders,
		COUNT(*) FILTER (WHERE order_status = 'confirmed') AS confirmed_orders,
		COUNT(*) FILTER (WHERE order_status = 'processing') AS processing_orders,
		COUNT(*) FILTER (WHERE order_status = 'shipped') AS shipped_orders,
		COUNT(*) FILTER (WHERE order_status = 'delivered') AS delivered_orders,
		COUNT(*) FILTER (WHERE order_status = 'cancelled') AS cancelled_orders,
		COUNT(*) FILTER (WHERE order_status = 'returned') AS returned_orders
	FROM orders;
`
	if err := r.db.WithContext(ctx).Raw(query).Scan(&overview).Error; err != nil {
		return nil, err
	}
	return &overview, nil
}
func (r *DashboardRepository) GetSales(ctx context.Context, filter models.DashboardFilter) ([]models.SalesData, error) {
	var sales []models.SalesData
	groupBy := filter.GroupBy
	switch groupBy {
	case "day", "week", "month", "year":
	default:
		return nil, fmt.Errorf("invalid sales grouping: %s", groupBy)
	}
	query := r.db.WithContext(ctx).
		Table("orders").
		Select(`
		DATE_TRUNC(?, placed_at) AS date,
		COALESCE(SUM(total_amount), 0) AS total_revenue,
		COUNT(*) AS total_orders
	`, groupBy).
		Where("payment_status IN ('paid', 'partially_refunded')")
	if !filter.StartDate.IsZero() {
		query = query.Where("placed_at >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query = query.Where("placed_at <= ?", filter.EndDate)
	}
	err := query.
		Group("DATE_TRUNC('" + groupBy + "', placed_at)").
		Order("date ASC").
		Scan(&sales).Error
	if err != nil {
		return nil, err
	}
	return sales, nil
}
func (r *DashboardRepository) GetOrderStats(ctx context.Context) (*models.OrderStats, error) {
	var stats models.OrderStats
	query := `
	SELECT
		COUNT(*) AS total_orders,
		COUNT(*) FILTER (WHERE order_status = 'pending') AS pending_orders,
		COUNT(*) FILTER (WHERE order_status = 'confirmed') AS confirmed_orders,
		COUNT(*) FILTER (WHERE order_status = 'processing') AS processing_orders,
		COUNT(*) FILTER (WHERE order_status = 'shipped') AS shipped_orders,
		COUNT(*) FILTER (WHERE order_status = 'delivered') AS delivered_orders,
		COUNT(*) FILTER (WHERE order_status = 'cancelled') AS cancelled_orders,
		COUNT(*) FILTER (WHERE order_status = 'returned') AS returned_orders
	FROM orders;
`
	if err := r.db.WithContext(ctx).Raw(query).Scan(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}
func (r *DashboardRepository) GetCustomerStats(ctx context.Context) (*models.CustomerStats, error) {
	var stats models.CustomerStats
	query := `
	SELECT
		COUNT(*) AS total_customers,
		COUNT(*) FILTER (
			WHERE is_blocked = TRUE
		) AS blocked_customers,
		COUNT(*) FILTER (
			WHERE email_verified_at IS NOT NULL
		) AS verified_customers,
		COUNT(*) FILTER (
			WHERE email_verified_at IS NULL
		) AS unverified_customers
	FROM users
	WHERE role = 'user';
`
	if err := r.db.WithContext(ctx).Raw(query).Scan(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}
func (r *DashboardRepository) GetTopProducts(ctx context.Context, limit int) ([]models.TopProduct, error) {
	var products []models.TopProduct
	if limit <= 0 {
		limit = 10
	}
	err := r.db.WithContext(ctx).
		Table("order_items").
		Select(`
		product_variants.product_id AS product_id,
		MAX(order_items.product_name) AS product_name,
		SUM(order_items.quantity) AS units_sold,
		COUNT(DISTINCT order_items.order_id) AS total_orders,
		SUM(order_items.subtotal) AS revenue
	`).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN product_variants ON product_variants.id = order_items.variant_id").
		Where("orders.payment_status IN ('paid', 'partially_refunded')").
		Where("orders.order_status NOT IN ('cancelled', 'returned')").
		Group("product_variants.product_id").
		Order("units_sold DESC").
		Limit(limit).
		Scan(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil

}
func (r *DashboardRepository) GetTopCategories(ctx context.Context, limit int) ([]models.TopCategory, error) {
	var categories []models.TopCategory
	if limit <= 0 {
		limit = 10
	}
	err := r.db.WithContext(ctx).
		Table("order_items").
		Select(`
		categories.id AS category_id,
		categories.name AS category_name,
		SUM(order_items.quantity) AS units_sold,
		SUM(order_items.subtotal) AS revenue
	`).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN product_variants ON product_variants.id = order_items.variant_id").
		Joins("JOIN products ON products.id = product_variants.product_id").
		Joins("JOIN categories ON categories.id = products.category_id").
		Where("orders.payment_status IN ('paid', 'partially_refunded')").
		Where("orders.order_status NOT IN ('cancelled', 'returned')").
		Group("categories.id, categories.name").
		Order("units_sold DESC").
		Limit(limit).
		Scan(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}
func (r *DashboardRepository) GetTopBrands(ctx context.Context, limit int) ([]models.TopBrand, error) {
	var brands []models.TopBrand
	if limit <= 0 {
		limit = 10
	}
	err := r.db.WithContext(ctx).
		Table("order_items").
		Select(`
		brands.id AS brand_id,
		brands.name AS brand_name,
		SUM(order_items.quantity) AS units_sold,
		SUM(order_items.subtotal) AS revenue
	`).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN product_variants ON product_variants.id = order_items.variant_id").
		Joins("JOIN products ON products.id = product_variants.product_id").
		Joins("JOIN brands ON brands.id = products.brand_id").
		Where("orders.payment_status IN ('paid', 'partially_refunded')").
		Where("orders.order_status NOT IN ('cancelled', 'returned')").
		Group("brands.id, brands.name").
		Order("units_sold DESC").
		Limit(limit).
		Scan(&brands).Error

	if err != nil {
		return nil, err
	}
	return brands, nil
}

var _ interfaces.DashboardRepository = (*DashboardRepository)(nil)
