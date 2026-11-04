package models

import "time"

// ReportFilter is shared by all report types.
type ReportFilter struct {
	StartDate time.Time
	EndDate   time.Time
	GroupBy   string
}

// SalesReport contains sales summary and breakdown.
type SalesReport struct {
	Summary   SalesReportSummary `json:"summary"`
	Breakdown []SalesReportRow   `json:"breakdown"`
}

type SalesReportSummary struct {
	TotalSales     float64 `json:"total_sales"`
	TotalOrders    int64   `json:"total_orders"`
	TotalUnitsSold int64   `json:"total_units_sold"`
	TotalDiscount  float64 `json:"total_discount"`
}

type SalesReportRow struct {
	Date      time.Time `json:"date"`
	Sales     float64   `json:"sales"`
	Orders    int64     `json:"orders"`
	UnitsSold int64     `json:"units_sold"`
}

// OrdersReport contains order summary and breakdown.
type OrdersReport struct {
	Summary   OrdersReportSummary `json:"summary"`
	Breakdown []OrdersReportRow   `json:"breakdown"`
}

type OrdersReportSummary struct {
	TotalOrders      int64 `json:"total_orders"`
	PendingOrders    int64 `json:"pending_orders"`
	ConfirmedOrders  int64 `json:"confirmed_orders"`
	ProcessingOrders int64 `json:"processing_orders"`
	ShippedOrders    int64 `json:"shipped_orders"`
	DeliveredOrders  int64 `json:"delivered_orders"`
	CancelledOrders  int64 `json:"cancelled_orders"`
	ReturnedOrders   int64 `json:"returned_orders"`
}

type OrdersReportRow struct {
	Date        time.Time `json:"date"`
	TotalOrders int64     `json:"total_orders"`
}

// ProductsReport contains product sales performance.
type ProductsReport struct {
	Summary   ProductsReportSummary `json:"summary"`
	Breakdown []ProductsReportRow   `json:"breakdown"`
}

type ProductsReportSummary struct {
	TotalProductsSold int64   `json:"total_products_sold"`
	TotalUnitsSold    int64   `json:"total_units_sold"`
	TotalRevenue      float64 `json:"total_revenue"`
}

type ProductsReportRow struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	UnitsSold   int64   `json:"units_sold"`
	TotalOrders int64   `json:"total_orders"`
	Revenue     float64 `json:"revenue"`
}
