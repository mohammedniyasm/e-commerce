package postgres

import (
	"context"
	"fmt"
	"strings"

	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GetSalesReport(ctx context.Context, filter models.ReportFilter) (*models.SalesReport, error) {
	groupBy, err := normalizeReportGroupBy(filter.GroupBy)
	if err != nil {
		return nil, err
	}
	where, args := buildReportOrderFilter("o", filter)
	report := &models.SalesReport{
		Breakdown: []models.SalesReportRow{},
	}
	summaryQuery := fmt.Sprintf(`
	WITH filtered_orders AS (
		SELECT o.id, o.total_amount, o.discount_amount, o.placed_at
		FROM orders o
		WHERE %s
	)
	SELECT
		COALESCE(SUM(total_amount), 0) AS total_sales,
		COUNT(*) AS total_orders,
		COALESCE(SUM(discount_amount), 0) AS total_discount,
		(
			SELECT COALESCE(SUM(oi.quantity), 0)
			FROM order_items oi
			JOIN filtered_orders fo ON fo.id = oi.order_id
		) AS total_units_sold
	FROM filtered_orders
`, where)
	if err := r.db.WithContext(ctx).
		Raw(summaryQuery, args...).
		Scan(&report.Summary).Error; err != nil {
		return nil, err
	}
	breakdownQuery := fmt.Sprintf(`
	WITH filtered_orders AS (
		SELECT o.id, o.total_amount, o.placed_at
		FROM orders o
		WHERE %s
	),
	order_groups AS (
		SELECT
			DATE_TRUNC('%s', placed_at) AS period,
			SUM(total_amount) AS sales,
			COUNT(*) AS orders
		FROM filtered_orders
		GROUP BY 1
	),
	item_groups AS (
		SELECT
			DATE_TRUNC('%s', fo.placed_at) AS period,
			SUM(oi.quantity) AS units_sold
		FROM filtered_orders fo
		JOIN order_items oi ON oi.order_id = fo.id
		GROUP BY 1
	)
	SELECT
		og.period AS date,
		og.sales,
		og.orders,
		COALESCE(ig.units_sold, 0) AS units_sold
	FROM order_groups og
	LEFT JOIN item_groups ig ON ig.period = og.period
	ORDER BY og.period ASC
`, where, groupBy, groupBy)
	if err := r.db.WithContext(ctx).
		Raw(breakdownQuery, args...).
		Scan(&report.Breakdown).Error; err != nil {
		return nil, err
	}
	return report, nil
}
func (r *ReportRepository) GetOrdersReport(ctx context.Context, filter models.ReportFilter) (*models.OrdersReport, error) {
	groupBy, err := normalizeReportGroupBy(filter.GroupBy)
	if err != nil {
		return nil, err
	}
	where, args := buildReportDateFilter("orders", filter)
	report := &models.OrdersReport{
		Breakdown: []models.OrdersReportRow{},
	}
	summaryQuery := fmt.Sprintf(`
	SELECT
		COUNT(*) AS total_orders,
		COUNT(*) FILTER (WHERE o.order_status = 'pending') AS pending_orders,
		COUNT(*) FILTER (WHERE o.order_status = 'confirmed') AS confirmed_orders,
		COUNT(*) FILTER (WHERE o.order_status = 'processing') AS processing_orders,
		COUNT(*) FILTER (WHERE o.order_status = 'shipped') AS shipped_orders,
		COUNT(*) FILTER (WHERE o.order_status = 'delivered') AS delivered_orders,
		COUNT(*) FILTER (WHERE o.order_status = 'cancelled') AS cancelled_orders,
		COUNT(*) FILTER (WHERE o.order_status = 'returned') AS returned_orders
	FROM orders o
	WHERE %s
`, where)
	if err := r.db.WithContext(ctx).
		Raw(summaryQuery, args...).
		Scan(&report.Summary).Error; err != nil {
		return nil, err
	}
	breakdownQuery := fmt.Sprintf(`
	SELECT
		DATE_TRUNC('%s', o.placed_at) AS date,
		COUNT(*) AS total_orders
	FROM orders o
	WHERE %s
	GROUP BY 1
	ORDER BY 1 ASC
`, groupBy, where)
	if err := r.db.WithContext(ctx).
		Raw(breakdownQuery, args...).
		Scan(&report.Breakdown).Error; err != nil {
		return nil, err
	}
	return report, nil
}
func (r *ReportRepository) GetProductsReport(ctx context.Context, filter models.ReportFilter) (*models.ProductsReport, error) {
	where, args := buildReportOrderFilter("o", filter)
	report := &models.ProductsReport{
		Breakdown: []models.ProductsReportRow{},
	}
	summaryQuery := fmt.Sprintf(`
	SELECT
		COUNT(DISTINCT pv.product_id) AS total_products_sold,
		COALESCE(SUM(oi.quantity), 0) AS total_units_sold,
		COALESCE(SUM(oi.subtotal), 0) AS total_revenue
	FROM orders o
	JOIN order_items oi ON oi.order_id = o.id
	JOIN product_variants pv ON pv.id = oi.variant_id
	WHERE %s
`, where)
	if err := r.db.WithContext(ctx).
		Raw(summaryQuery, args...).
		Scan(&report.Summary).Error; err != nil {
		return nil, err
	}
	breakdownQuery := fmt.Sprintf(`
	SELECT
		pv.product_id AS product_id,
		oi.product_name AS product_name,
		SUM(oi.quantity) AS units_sold,
		COUNT(DISTINCT o.id) AS total_orders,
		SUM(oi.subtotal) AS revenue
	FROM orders o
	JOIN order_items oi ON oi.order_id = o.id
	JOIN product_variants pv ON pv.id = oi.variant_id
	WHERE %s
	GROUP BY pv.product_id, oi.product_name
	ORDER BY units_sold DESC
`, where)
	if err := r.db.WithContext(ctx).
		Raw(breakdownQuery, args...).
		Scan(&report.Breakdown).Error; err != nil {
		return nil, err
	}
	return report, nil
}
func normalizeReportGroupBy(groupBy string) (string, error) {
	groupBy = strings.ToLower(strings.TrimSpace(groupBy))
	if groupBy == "" {
		return "day", nil
	}
	switch groupBy {
	case "day", "week", "month", "year":
		return groupBy, nil
	default:
		return "", fmt.Errorf("invalid report grouping: %s", groupBy)
	}
}
func buildReportOrderFilter(alias string, filter models.ReportFilter) (string, []interface{}) {
	conditions := []string{
		alias + ".payment_status IN ('paid', 'partially_refunded')",
		alias + ".order_status NOT IN ('cancelled', 'returned')",
	}
	var args []interface{}
	if !filter.StartDate.IsZero() {
		conditions = append(
			conditions,
			alias+".placed_at >= ?",
		)
		args = append(args, filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		conditions = append(
			conditions,
			alias+".placed_at <= ?",
		)
		args = append(args, filter.EndDate)
	}
	return strings.Join(conditions, " AND "), args
}
func buildReportDateFilter(alias string, filter models.ReportFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	if !filter.StartDate.IsZero() {
		args = append(args, filter.StartDate)
		conditions = append(
			conditions,
			fmt.Sprintf("%s.placed_at >= ?", alias),
		)
	}
	if !filter.EndDate.IsZero() {
		args = append(args, filter.EndDate)
		conditions = append(
			conditions,
			fmt.Sprintf("%s.placed_at <= ?", alias),
		)
	}
	if len(conditions) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

var _ interfaces.ReportRepository = (*ReportRepository)(nil)
