package response

import (
	// "ecommerce/internal/delivery/http/dto/response"

	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PlaceOrderResponse struct {
	OrderID                uuid.UUID  `json:"order_id"`
	OrderNumber            string     `json:"order_number"`
	OrderStatus            string     `json:"order_status"`
	PaymentStatus          string     `json:"payment_status"`
	Subtotal               float64    `json:"subtotal"`
	DiscountAmount         float64    `json:"discount_amount"`
	ShippingAmount         float64    `json:"shipping_amount"`
	TotalAmount            float64    `json:"total_amount"`
	EstimatedDeliveryStart *time.Time `json:"estimated_delivery_start,omitempty"`
	EstimatedDeliveryEnd   *time.Time `json:"estimated_delivery_end,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
}

type OrderItemResponse struct {
	ID          uint            `json:"id"`
	VariantID   uint            `json:"variant_id"`
	ProductName string          `json:"product_name"`
	ImageURL    string          `json:"image_url"`
	VariantInfo json.RawMessage `json:"variant_info"`
	UnitPrice   float64         `json:"unit_price"`
	Quantity    int             `json:"quantity"`
	Subtotal    float64         `json:"subtotal"`
}

type OrderAddressResponse struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Pincode      string `json:"pincode"`
}

type OrderResponse struct {
	ID                     uuid.UUID             `json:"id"`
	OrderNumber            string                `json:"order_number"`
	OrderStatus            string                `json:"order_status"`
	PaymentStatus          string                `json:"payment_status"`
	Subtotal               float64               `json:"subtotal"`
	DiscountAmount         float64               `json:"discount_amount"`
	ShippingAmount         float64               `json:"shipping_amount"`
	TotalAmount            float64               `json:"total_amount"`
	EstimatedDeliveryStart *time.Time            `json:"estimated_delivery_start,omitempty"`
	EstimatedDeliveryEnd   *time.Time            `json:"estimated_delivery_end,omitempty"`
	Items                  []OrderItemResponse   `json:"items"`
	Address                *OrderAddressResponse `json:"address,omitempty"`
	PlacedAt               time.Time             `json:"placed_at"`
	CreatedAt              time.Time             `json:"created_at"`
}

type OrderActionResponse struct {
	OrderID uuid.UUID `json:"order_id"`
	Message string    `json:"message"`
}

type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int64           `json:"total_pages"`
}
