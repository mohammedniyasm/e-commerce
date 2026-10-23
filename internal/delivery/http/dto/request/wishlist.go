package request

type AddWishlistProductRequest struct {
	ProductID int64 `json:"product_id" binding:"required,min=1"`
}

type MergeGuestWishlistRequest struct {
	ProductIDs []int64 `json:"product_ids" binding:"required,min=1,max=100,dive,gt=0"`
}
