package response

import (
	"ecommerce/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type ReturnRequestResponse struct {
	ID                 uint      `json:"id"`
	OrderID            uuid.UUID `json:"order_id"`
	OrderItemID        uint      `json:"order_item_id"`
	RequestType        string    `json:"request_type"`
	RequestedVariantID *uint     `json:"requested_variant_id,omitempty"`
	Quantity           int       `json:"quantity"`
	Reason             string    `json:"reason"`
	Status             string    `json:"status"`
	AdminNote          *string   `json:"admin_note,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func ToReturnRequestResponse(request *models.ReturnRequest,
) ReturnRequestResponse {
	return ReturnRequestResponse{
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
}
