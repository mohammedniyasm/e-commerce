package models

import "time"

type DashboardOverview struct {
	TotalRevenue     float64 `gorm:"column:total_revenue"`
	TotalOrders      int64   `gorm:"column:total_orders"`
	TotalCustomers   int64   `gorm:"column:total_customers"`
	TotalProducts    int64   `gorm:"column:total_products"`
	PendingOrders    int64   `gorm:"column:pending_orders"`
	ConfirmedOrders  int64   `gorm:"column:confirmed_orders"`
	ProcessingOrders int64   `gorm:"column:processing_orders"`
	ShippedOrders    int64   `gorm:"column:shipped_orders"`
	DeliveredOrders  int64   `gorm:"column:delivered_orders"`
	CancelledOrders  int64   `gorm:"column:cancelled_orders"`
	ReturnedOrders   int64   `gorm:"column:returned_orders"`
}

type DashboardFilter struct {
	StartDate time.Time
	EndDate   time.Time
	GroupBy   string // day, week, month, year
}

type SalesData struct {
	Date         time.Time `json:"date"`
	TotalRevenue float64   `json:"total_revenue"`
	TotalOrders  int64     `json:"total_orders"`
}

type OrderStats struct {
	TotalOrders      int64 `json:"total_orders"`
	PendingOrders    int64 `json:"pending_orders"`
	ConfirmedOrders  int64 `json:"confirmed_orders"`
	ProcessingOrders int64 `json:"processing_orders"`
	ShippedOrders    int64 `json:"shipped_orders"`
	DeliveredOrders  int64 `json:"delivered_orders"`
	CancelledOrders  int64 `json:"cancelled_orders"`
	ReturnedOrders   int64 `json:"returned_orders"`
}

type CustomerStats struct {
	TotalCustomers      int64 `json:"total_customers"`
	BlockedCustomers    int64 `json:"blocked_customers"`
	VerifiedCustomers   int64 `json:"verified_customers"`
	UnverifiedCustomers int64 `json:"unverified_customers"`
}

type TopProduct struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	UnitsSold   int64   `json:"units_sold"`
	TotalOrders int64   `json:"total_orders"`
	Revenue     float64 `json:"revenue"`
}

type TopCategory struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	UnitsSold    int64   `json:"units_sold"`
	Revenue      float64 `json:"revenue"`
}

type TopBrand struct {
	BrandID   uint    `json:"brand_id"`
	BrandName string  `json:"brand_name"`
	UnitsSold int64   `json:"units_sold"`
	Revenue   float64 `json:"revenue"`
}
