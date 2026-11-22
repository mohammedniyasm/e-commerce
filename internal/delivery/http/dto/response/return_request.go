package response

import (
	"ecommerce/internal/domain/models"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ReturnRequestOrderResponse struct {
	ID            uuid.UUID `json:"id"`
	OrderNumber   string    `json:"order_number"`
	OrderStatus   string    `json:"order_status"`
	PaymentStatus string    `json:"payment_status"`
	UserID        uint      `json:"user_id"`
}

type ReturnRequestOrderItemResponse struct {
	ID          uint            `json:"id"`
	VariantID   uint            `json:"variant_id"`
	ProductName string          `json:"product_name"`
	ImageURL    string          `json:"image_url"`
	VariantInfo json.RawMessage `json:"variant_info"`
	UnitPrice   float64         `json:"unit_price"`
	Quantity    int             `json:"quantity"`
	Subtotal    float64         `json:"subtotal"`
}

type ReturnRequestVariantResponse struct {
	ID           uint    `json:"id"`
	ProductID    uint    `json:"product_id"`
	Size         *string `json:"size,omitempty"`
	Color        *string `json:"color,omitempty"`
	SKU          string  `json:"sku"`
	ImageURL     string  `json:"image_url,omitempty"`
	SellingPrice float64 `json:"selling_price"`
	Stock        int     `json:"stock"`
}

type ReturnRequestResponse struct {
	ID                 uint                            `json:"id"`
	OrderID            uuid.UUID                       `json:"order_id"`
	OrderItemID        uint                            `json:"order_item_id"`
	RequestType        string                          `json:"request_type"`
	RequestedVariantID *uint                           `json:"requested_variant_id,omitempty"`
	Quantity           int                             `json:"quantity"`
	Reason             string                          `json:"reason"`
	Status             string                          `json:"status"`
	AdminNote          *string                         `json:"admin_note,omitempty"`
	CreatedAt          time.Time                       `json:"created_at"`
	UpdatedAt          time.Time                       `json:"updated_at"`
	Order              *ReturnRequestOrderResponse     `json:"order,omitempty"`
	OrderItem          *ReturnRequestOrderItemResponse `json:"order_item,omitempty"`
	RequestedVariant   *ReturnRequestVariantResponse   `json:"requested_variant,omitempty"`
}

func ToReturnRequestResponse(request *models.ReturnRequest,
) ReturnRequestResponse {
	result := ReturnRequestResponse{
		ID:                 request.ID,
		OrderID:            request.OrderID,
		OrderItemID:        request.OrderItemID,
		RequestType:        request.RequestType,
		RequestedVariantID: request.RequestedVariantID,
		Quantity:           request.Quantity,
		Reason:             request.Reason,
		Status:             request.Status,
		AdminNote:          request.AdminNote,
		CreatedAt:          request.CreatedAt,
		UpdatedAt:          request.UpdatedAt,
	}
	if request.Order.ID != uuid.Nil {
		result.Order = &ReturnRequestOrderResponse{
			ID:            request.Order.ID,
			OrderNumber:   request.Order.OrderNumber,
			OrderStatus:   request.Order.OrderStatus,
			PaymentStatus: request.Order.PaymentStatus,
			UserID:        request.Order.UserID,
		}
	}
	if request.OrderItem.ID != 0 {
		result.OrderItem = &ReturnRequestOrderItemResponse{
			ID:          request.OrderItem.ID,
			VariantID:   request.OrderItem.VariantID,
			ProductName: request.OrderItem.ProductName,
			ImageURL:    request.OrderItem.ImageURL,
			VariantInfo: request.OrderItem.VariantInfo,
			UnitPrice:   request.OrderItem.UnitPrice,
			Quantity:    request.OrderItem.Quantity,
			Subtotal:    request.OrderItem.Subtotal,
		}
	}
	if request.RequestedVariant != nil {
		result.RequestedVariant = &ReturnRequestVariantResponse{
			ID:           request.RequestedVariant.ID,
			ProductID:    request.RequestedVariant.ProductID,
			Size:         request.RequestedVariant.Size,
			Color:        request.RequestedVariant.Color,
			SKU:          request.RequestedVariant.SKU,
			ImageURL:     requestedVariantImageURL(request.RequestedVariant),
			SellingPrice: request.RequestedVariant.SellingPrice,
			Stock:        request.RequestedVariant.Stock,
		}
	}
	return result
}

func requestedVariantImageURL(variant *models.ProductVariant) string {
	for _, image := range variant.Images {
		if image.IsPrimary {
			return image.ImageURL
		}
	}
	if len(variant.Images) > 0 {
		return variant.Images[0].ImageURL
	}
	for _, image := range variant.Product.Images {
		if image.VariantID == nil && image.IsPrimary {
			return image.ImageURL
		}
	}
	for _, image := range variant.Product.Images {
		if image.VariantID == nil {
			return image.ImageURL
		}
	}
	return ""
}
