package response

import "ecommerce/internal/usecase/interfaces"

type GuestCartItemValidationResponse struct {
	VariantID          int64   `json:"variant_id"`
	RequestedQuantity  int     `json:"requested_quantity"`
	AvailableStock     int     `json:"available_stock"`
	CurrentPrice       float64 `json:"current_price"`
	IsAvailable        bool    `json:"is_available"`
	AvailabilityReason string  `json:"availability_reason,omitempty"`
}

func ToGuestCartItemValidationResponse(
	result interfaces.GuestCartItemResult,
) GuestCartItemValidationResponse {
	return GuestCartItemValidationResponse{
		VariantID:          result.VariantID,
		RequestedQuantity:  result.RequestedQuantity,
		AvailableStock:     result.AvailableStock,
		CurrentPrice:       result.CurrentPrice,
		IsAvailable:        result.IsAvailable,
		AvailabilityReason: result.AvailabilityReason,
	}
}

func ToGuestCartValidationResponses(
	results []interfaces.GuestCartItemResult,
) []GuestCartItemValidationResponse {
	responses := make(
		[]GuestCartItemValidationResponse,
		0,
		len(results),
	)

	for _, result := range results {
		responses = append(
			responses,
			ToGuestCartItemValidationResponse(result),
		)
	}

	return responses
}
