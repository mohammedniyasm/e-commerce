package request

type CreateReturnRequestDTO struct {
	OrderID            string `json:"order_id" binding:"required,uuid"`
	OrderItemID        uint   `json:"order_item_id" binding:"required,gt=0"`
	RequestType        string `json:"request_type" binding:"required,oneof=return exchange"`
	RequestedVariantID *uint  `json:"requested_variant_id,omitempty" binding:"omitempty,gt=0"`
	Quantity           int    `json:"quantity" binding:"required,gt=0"`
	Reason             string `json:"reason" binding:"required,min=5,max=1000"`
}

type UpdateReturnRequestStatusDTO struct {
	Status    string  `json:"status" binding:"required,oneof=approved rejected pickup_scheduled in_transit received inspection_pending inspection_passed inspection_failed closed"`
	AdminNote *string `json:"admin_note,omitempty" binding:"omitempty,max=1000"`
}
