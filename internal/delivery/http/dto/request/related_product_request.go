package request

type RelatedProductsQueryRequest struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=20"`
}
