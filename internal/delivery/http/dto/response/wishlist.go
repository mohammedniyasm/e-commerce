package response

import (
	"ecommerce/internal/domain/models"
	"time"
)

type WishlistItemResponse struct {
	ID                 uint      `json:"id"`
	ProductID          uint      `json:"product_id"`
	ProductName        string    `json:"product_name"`
	ProductSlug        string    `json:"product_slug"`
	ShortDescription   *string   `json:"short_description,omitempty"`
	CategoryName       string    `json:"category_name"`
	BrandName          string    `json:"brand_name"`
	ImageURL           *string   `json:"image_url,omitempty"`
	IsAvailable        bool      `json:"is_available"`
	AvailabilityReason string    `json:"availability_reason,omitempty"`
	AddedAt            time.Time `json:"added_at"`
}

type WishlistResponse struct {
	ID        uint                   `json:"id"`
	Items     []WishlistItemResponse `json:"items"`
	ItemCount int                    `json:"item_count"`
}

func ToWishlistItemResponse(item *models.WishlistItem) WishlistItemResponse {
	product := item.Product
	isAvailable := true
	availabilityReason := ""
	if product.DeletedAt != nil ||
		product.Category.DeletedAt != nil ||
		product.Brand.DeletedAt != nil ||
		!product.IsActive ||
		!product.IsListed ||
		product.ID == 0 ||
		product.Category.ID == 0 ||
		product.Brand.ID == 0 ||
		!product.Category.IsActive ||
		!product.Brand.IsActive {
		isAvailable = false
		availabilityReason = "product_unavailable"
	}
	productName := product.Name
	if product.ID == 0 {
		productName = "Unavailable product"
	}
	var imageURL *string
	for _, image := range product.Images {
		if image.IsPrimary {
			imageURL = &image.ImageURL
			break
		}
	}
	return WishlistItemResponse{
		ID:                 item.ID,
		ProductID:          item.ProductID,
		ProductName:        productName,
		ProductSlug:        product.Slug,
		ShortDescription:   product.ShortDescription,
		CategoryName:       product.Category.Name,
		BrandName:          product.Brand.Name,
		ImageURL:           imageURL,
		IsAvailable:        isAvailable,
		AvailabilityReason: availabilityReason,
		AddedAt:            item.CreatedAt,
	}
}
func ToWishlistResponse(wishlist *models.Wishlist) WishlistResponse {
	result := WishlistResponse{
		Items: make([]WishlistItemResponse, 0),
	}
	if wishlist == nil {
		return result
	}
	result.ID = wishlist.ID
	result.Items = make([]WishlistItemResponse, 0, len(wishlist.Items))
	for _, item := range wishlist.Items {
		result.Items = append(result.Items, ToWishlistItemResponse(&item))
	}
	result.ItemCount = len(result.Items)
	return result
}
