package request

type GuestCartValidationItemRequest struct {
	VariantID int64 `json:"variant_id" binding:"required,min=1"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

type ValidateGuestCartRequest struct {
	Items []GuestCartValidationItemRequest `json:"items" binding:"max=100,dive"`
}
