package response

import "ecommerce/internal/domain/models"

type RelatedProductResponse struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Slug             string  `json:"slug"`
	ShortDescription *string `json:"short_description,omitempty"`
	CategoryName     string  `json:"category_name"`
	BrandName        string  `json:"brand_name"`
	StartingPrice    float64 `json:"starting_price"`
	PrimaryImageURL  *string `json:"primary_image_url,omitempty"`
	IsInStock        bool    `json:"is_in_stock"`
}

func ToRelatedProductResponse(product models.Product) RelatedProductResponse {
	var startingPrice float64
	var hasPrice bool
	var isInStock bool

	for _, variant := range product.Variants {
		if !variant.IsActive {
			continue
		}

		if !hasPrice || variant.SellingPrice < startingPrice {
			startingPrice = variant.SellingPrice
			hasPrice = true
		}

		if variant.Stock > 0 {
			isInStock = true
		}
	}

	var primaryImageURL *string

	for _, image := range product.Images {
		if image.IsPrimary {
			imageURL := image.ImageURL
			primaryImageURL = &imageURL
			break
		}
	}

	return RelatedProductResponse{
		ID:               product.ID,
		Name:             product.Name,
		Slug:             product.Slug,
		ShortDescription: product.ShortDescription,
		CategoryName:     product.Category.Name,
		BrandName:        product.Brand.Name,
		StartingPrice:    startingPrice,
		PrimaryImageURL:  primaryImageURL,
		IsInStock:        isInStock,
	}
}
func ToRelatedProductsResponse(products []models.Product) []RelatedProductResponse {
	responses := make([]RelatedProductResponse, 0, len(products))

	for _, product := range products {
		responses = append(
			responses,
			ToRelatedProductResponse(product),
		)
	}

	return responses
}
