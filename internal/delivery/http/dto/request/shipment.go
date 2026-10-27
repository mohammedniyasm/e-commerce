package request

import "github.com/google/uuid"

type CreateShipmentRequest struct {
	OrderID        uuid.UUID `json:"order_id" binding:"required"`
	CourierName    *string   `json:"courier_name"`
	TrackingNumber *string   `json:"tracking_number"`
}

type UpdateShipmentRequest struct {
	CourierName    *string `json:"courier_name"`
	TrackingNumber *string `json:"tracking_number"`
}

type UpdateShippingStatusRequest struct {
	Status      string  `json:"status" binding:"required"`
	Description *string `json:"description"`
	Location    *string `json:"location"`
}
