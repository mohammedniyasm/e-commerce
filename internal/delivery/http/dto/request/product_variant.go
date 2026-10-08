package request

type CreateProductVariantRequest struct {
	Size         *string `json:"size"`
	Color        *string `json:"color"`
	SKU          string  `json:"sku" binding:"required"`
	MRP          float64 `json:"mrp" binding:"required"`
	SellingPrice float64 `json:"selling_price" binding:"required"`
	Stock        int     `json:"stock"`
}

type UpdateProductVariantRequest struct {
	Size         *string `json:"size"`
	Color        *string `json:"color"`
	SKU          string  `json:"sku" binding:"required"`
	MRP          float64 `json:"mrp" binding:"required"`
	SellingPrice float64 `json:"selling_price" binding:"required"`
	Stock        int     `json:"stock"`
	IsActive     *bool    `json:"is_active"`
}
