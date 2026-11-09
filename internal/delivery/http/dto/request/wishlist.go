package request

type CreateWishlistRequest struct {
	Name string `json:"name" binding:"required,min=2,max=50"`
}

type UpdateWishlistRequest struct {
	Name string `json:"name" binding:"required,min=2,max=50"`
}

type AddProductToWishlistRequest struct {
	WishlistName string `json:"wishlist_name" binding:"required,min=2,max=50"`
	ProductID    uint   `json:"product_id" binding:"required"`
}

type MergeWishlistItemsRequest struct {
	WishlistName string `json:"wishlist_name" binding:"required,min=2,max=50"`
	ProductIDs   []uint `json:"product_ids" binding:"required,min=1,dive,gt=0"`
}
