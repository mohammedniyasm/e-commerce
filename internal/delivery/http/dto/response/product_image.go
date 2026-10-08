package response

import (
	"time"

	"ecommerce/internal/domain/models"
)

type ProductImageResponse struct {
	ID           uint      `json:"id"`
	ProductID    uint      `json:"product_id"`
	VariantID    *uint     `json:"variant_id"`
	ImageURL     string    `json:"image_url"`
	DisplayOrder int       `json:"display_order"`
	IsPrimary    bool      `json:"is_primary"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToProductImageResponse(
	image *models.ProductImage,
) ProductImageResponse {
	return ProductImageResponse{
		ID:           image.ID,
		ProductID:    image.ProductID,
		VariantID:    image.VariantID,
		ImageURL:     image.ImageURL,
		DisplayOrder: image.DisplayOrder,
		IsPrimary:    image.IsPrimary,
		CreatedAt:    image.CreatedAt,
		UpdatedAt:    image.UpdatedAt,
	}
}
