package response

import (
	"time"

	"ecommerce/internal/domain/models"
)

type ProductResponse struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Slug             string  `json:"slug"`
	ShortDescription *string `json:"short_description"`
	Description      *string `json:"description"`

	CategoryID uint `json:"category_id"`
	BrandID    uint `json:"brand_id"`

	IsActive bool `json:"is_active"`
	IsListed bool `json:"is_listed"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"total_pages"`
}

func ToProductResponse(product *models.Product) ProductResponse {
	return ProductResponse{
		ID:               product.ID,
		Name:             product.Name,
		Slug:             product.Slug,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		CategoryID:       product.CategoryID,
		BrandID:          product.BrandID,
		IsActive:         product.IsActive,
		IsListed:         product.IsListed,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	}
}
