package response

import (
	"ecommerce/internal/domain/models"
	"time"
)

type StoreProductDetailCategory struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type StoreProductDetailBrand struct {
	ID   uint    `json:"id"`
	Name string  `json:"name"`
	Logo *string `json:"logo"`
}

type StoreProductDetailImage struct {
	ID           uint   `json:"id"`
	ImageURL     string `json:"image_url"`
	DisplayOrder int    `json:"display_order"`
	IsPrimary    bool   `json:"is_primary"`
}

type StoreProductVariantDetail struct {
	ID           uint                      `json:"id"`
	Size         *string                   `json:"size"`
	Color        *string                   `json:"color"`
	SKU          string                    `json:"sku"`
	MRP          float64                   `json:"mrp"`
	SellingPrice float64                   `json:"selling_price"`
	Stock        int                       `json:"stock"`
	InStock      bool                      `json:"in_stock"`
	Images       []StoreProductDetailImage `json:"images"`
}

type StoreProductDetailResponse struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Slug             string  `json:"slug"`
	ShortDescription *string `json:"short_description"`
	Description      *string `json:"description"`

	Category StoreProductDetailCategory `json:"category"`
	Brand    StoreProductDetailBrand    `json:"brand"`

	Images   []StoreProductDetailImage   `json:"images"`
	Variants []StoreProductVariantDetail `json:"variants"`

	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
	InStock  bool    `json:"in_stock"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToStoreProductDetailResponse(product *models.Product) StoreProductDetailResponse {
	productImages := make([]StoreProductDetailImage,0,len(product.Images))
	for _, image := range product.Images {
		productImages = append(
			productImages,
			StoreProductDetailImage{
				ID:           image.ID,
				ImageURL:     image.ImageURL,
				DisplayOrder: image.DisplayOrder,
				IsPrimary:    image.IsPrimary,
			},
		)
	}

	variants := make([]StoreProductVariantDetail,0,len(product.Variants))

	var minPrice float64
	var maxPrice float64
	var productInStock bool

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

		variantInStock := variant.Stock > 0

		if variantInStock {
			productInStock = true
		}
		variantImages := make([]StoreProductDetailImage,0,len(variant.Images))

		for _, image := range variant.Images {
			variantImages = append(
				variantImages,
				StoreProductDetailImage{
					ID:           image.ID,
					ImageURL:     image.ImageURL,
					DisplayOrder: image.DisplayOrder,
					IsPrimary:    image.IsPrimary,
				},
			)
		}

		variants = append(
			variants,
			StoreProductVariantDetail{
				ID:           variant.ID,
				Size:         variant.Size,
				Color:        variant.Color,
				SKU:          variant.SKU,
				MRP:          variant.MRP,
				SellingPrice: variant.SellingPrice,
				Stock:        variant.Stock,
				InStock:      variantInStock,
				Images:       variantImages,
			},
		)
	}

	return StoreProductDetailResponse{
		ID:               product.ID,
		Name:             product.Name,
		Slug:             product.Slug,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,

		Category: StoreProductDetailCategory{
			ID:   uint(product.Category.ID),
			Name: product.Category.Name,
		},

		Brand: StoreProductDetailBrand{
			ID:   product.Brand.ID,
			Name: product.Brand.Name,
			Logo: product.Brand.Logo,
		},

		Images:   productImages,
		Variants: variants,

		MinPrice: minPrice,
		MaxPrice: maxPrice,
		InStock:  productInStock,

		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}
