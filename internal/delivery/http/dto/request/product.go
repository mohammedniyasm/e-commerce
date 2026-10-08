package request

type CreateProductRequest struct {
	Name             string  `json:"name" binding:"required"`
	ShortDescription *string `json:"short_description"`
	Description      *string `json:"description"`

	CategoryID int64 `json:"category_id" binding:"required"`
	BrandID    int64 `json:"brand_id" binding:"required"`
}

type UpdateProductRequest struct {
	Name             string  `json:"name" binding:"required"`
	ShortDescription *string `json:"short_description"`
	Description      *string `json:"description"`

	CategoryID int64 `json:"category_id" binding:"required"`
	BrandID    int64 `json:"brand_id" binding:"required"`
}

type ListProductRequest struct {
	Search     string `form:"search"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	IsActive   *bool  `form:"is_active"`
	IsListed   *bool  `form:"is_listed"`
	CategoryID *int64 `form:"category_id"`
	BrandID    *int64 `form:"brand_id"`
}
