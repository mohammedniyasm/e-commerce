package request

type UploadProductImagesRequest struct {
	VariantID string `form:"variant_id"`
}

type UpdateProductImageRequest struct {
	VariantID    *int64 `json:"variant_id"`
	DisplayOrder *int    `json:"display_order"`
	IsPrimary    *bool   `json:"is_primary"`
}
