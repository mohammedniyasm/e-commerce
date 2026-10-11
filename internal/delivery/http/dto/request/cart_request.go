package request

type AddCartItemRequest struct {
	VariantID int64 `json:"variant_id" binding:"required,min=1"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemQuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}
type MergeCartItemRequest struct {
	Items []AddCartItemRequest `json:"items" binding:"required,min=1,max=100,dive"`
}
