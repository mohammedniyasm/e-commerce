package request

type CreateReviewRequest struct {
	OrderItemID uint    `json:"order_item_id" binding:"required,gt=0"`
	Rating      int16   `json:"rating" binding:"required,min=1,max=5"`
	Comment     *string `json:"comment"`
}

type UpdateReviewRequest struct {
	Rating  int16   `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment"`
}

type UpdateReviewVisibilityRequest struct {
	IsVisible *bool `json:"is_visible" binding:"required"`
}
