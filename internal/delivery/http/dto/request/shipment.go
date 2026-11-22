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

type ListShipmentsRequest struct {
	Search          string `form:"search" binding:"max=255"`
	ShippingStatus  string `form:"shipping_status"`
	Type            string `form:"type"`
	CourierName     string `form:"courier_name" binding:"max=255"`
	OrderID         string `form:"order_id"`
	ReturnRequestID string `form:"return_request_id"`
	CreatedFrom     string `form:"created_from"`
	CreatedTo       string `form:"created_to"`
	Sort            string `form:"sort"`
	Limit           int    `form:"limit,default=10"`
	Page            int    `form:"page,default=1"`
}
