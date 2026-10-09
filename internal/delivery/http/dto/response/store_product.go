package response

import (
	"ecommerce/internal/domain/models"
	"time"
)

type StoreProductImageResponse struct {
	ID       uint   `json:"id"`
	ImageURL string `json:"image_url"`
}

type StoreProductResponse struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Slug             string  `json:"slug"`
	ShortDescription *string `json:"short_description"`

	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`

	BrandID   uint   `json:"brand_id"`
	BrandName string `json:"brand_name"`

	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`

	PrimaryImage *StoreProductImageResponse `json:"primary_image"`

	InStock bool `json:"in_stock"`

	CreatedAt time.Time `json:"created_at"`
}

type StoreProductListResponse struct {
	Products   []StoreProductResponse `json:"products"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	Total      int64                  `json:"total"`
	TotalPages int                    `json:"total_pages"`
}

func ToStoreProductResponse(product *models.Product) StoreProductResponse {

	var minPrice float64
	var maxPrice float64
	var inStock bool

	for i, variant := range product.Variants {
		if i == 0 {
			minPrice = variant.SellingPrice
			maxPrice = variant.SellingPrice
		} else {
			if variant.SellingPrice < minPrice {
				minPrice = variant.SellingPrice
			}

			if variant.SellingPrice > maxPrice {
				maxPrice = variant.SellingPrice
			}
		}

		if variant.Stock > 0 {
			inStock = true
		}
	}

	var primaryImage *StoreProductImageResponse

	if len(product.Images) > 0 {
		primaryImage = &StoreProductImageResponse{
			ID:       product.Images[0].ID,
			ImageURL: product.Images[0].ImageURL,
		}
	}

	return StoreProductResponse{
		ID:               product.ID,
		Name:             product.Name,
		Slug:             product.Slug,
		ShortDescription: product.ShortDescription,

		CategoryID:   product.CategoryID,
		CategoryName: product.Category.Name,

		BrandID:   product.BrandID,
		BrandName: product.Brand.Name,

		MinPrice: minPrice,
		MaxPrice: maxPrice,

		PrimaryImage: primaryImage,
		InStock:      inStock,

		CreatedAt: product.CreatedAt,
	}
}
