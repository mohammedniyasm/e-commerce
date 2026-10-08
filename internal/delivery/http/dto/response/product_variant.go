package response

import (
	"time"

	"ecommerce/internal/domain/models"
)

type ProductVariantResponse struct {
	ID           uint      `json:"id"`
	ProductID    uint      `json:"product_id"`
	Size         *string   `json:"size"`
	Color        *string   `json:"color"`
	SKU          string    `json:"sku"`
	MRP          float64   `json:"mrp"`
	SellingPrice float64   `json:"selling_price"`
	Stock        int       `json:"stock"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToProductVariantResponse(
	variant *models.ProductVariant,
) ProductVariantResponse {
	return ProductVariantResponse{
		ID:           variant.ID,
		ProductID:    variant.ProductID,
		Size:         variant.Size,
		Color:        variant.Color,
		SKU:          variant.SKU,
		MRP:          variant.MRP,
		SellingPrice: variant.SellingPrice,
		Stock:        variant.Stock,
		IsActive:     variant.IsActive,
		CreatedAt:    variant.CreatedAt,
		UpdatedAt:    variant.UpdatedAt,
	}
}
