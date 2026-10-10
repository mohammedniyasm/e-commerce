package request

type ListStoreProductsRequest struct {
	Search     string   `form:"search"`
	CategoryID *int64   `form:"category_id"`
	BrandID    *int64   `form:"brand_id"`
	MinPrice   *float64 `form:"min_price"`
	MaxPrice   *float64 `form:"max_price"`
	Sort       string   `form:"sort"`
	Page       int      `form:"page"`
	Limit      int      `form:"limit"`
}
