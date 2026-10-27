package response

import (
	"time"

	"github.com/google/uuid"
)

type ShipmentResponse struct {
	ID             uint       `json:"id"`
	OrderID        uuid.UUID  `json:"order_id"`
	CourierName    *string    `json:"courier_name,omitempty"`
	TrackingNumber *string    `json:"tracking_number,omitempty"`
	ShippingStatus string     `json:"shipping_status"`
	ShippedAt      *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ShipmentTrackingHistoryResponse struct {
	ID             uint      `json:"id"`
	ShipmentID     uint      `json:"shipment_id"`
	ShippingStatus string    `json:"shipping_status"`
	Description    *string   `json:"description,omitempty"`
	Location       *string   `json:"location,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type ShipmentTrackingHistoryListResponse struct {
	ShipmentID uint                              `json:"shipment_id"`
	History    []ShipmentTrackingHistoryResponse `json:"history"`
	Total      int                               `json:"total"`
}

type ShipmentMessageResponse struct {
	ShipmentID uint   `json:"shipment_id"`
	Message    string `json:"message"`
}
type CustomerShipmentResponse struct {
	ID             uint       `json:"id"`
	OrderID        uuid.UUID  `json:"order_id"`
	CourierName    *string    `json:"courier_name,omitempty"`
	TrackingNumber *string    `json:"tracking_number,omitempty"`
	ShippingStatus string     `json:"shipping_status"`
	ShippedAt      *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CustomerShipmentHistoryItemResponse struct {
	ID             uint      `json:"id"`
	ShippingStatus string    `json:"shipping_status"`
	Description    *string   `json:"description,omitempty"`
	Location       *string   `json:"location,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type CustomerShipmentHistoryResponse struct {
	ShipmentID uint                                  `json:"shipment_id"`
	History    []CustomerShipmentHistoryItemResponse `json:"history"`
	Total      int                                   `json:"total"`
}
