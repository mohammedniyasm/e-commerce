package response

import (
	"ecommerce/internal/domain/models"
	"time"
)

type WishlistResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	ItemCount int64     `json:"item_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type WishlistProductResponse struct {
	ID                 uint      `json:"id"`
	ProductID          uint      `json:"product_id"`
	ProductName        string    `json:"product_name"`
	ProductSlug        string    `json:"product_slug"`
	ShortDescription   string    `json:"short_description"`
	Description        string    `json:"description"`
	ImageURL           string    `json:"image_url"`
	IsAvailable        bool      `json:"is_available"`
	AvailabilityReason string    `json:"availability_reason,omitempty"`
	AddedAt            time.Time `json:"added_at"`
}
type WishlistDetailsResponse struct {
	ID        uint                      `json:"id"`
	Name      string                    `json:"name"`
	Items     []WishlistProductResponse `json:"items"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}
type MergeWishlistItemsResponse struct {
	Wishlist WishlistDetailsResponse `json:"wishlist"`
	Accepted uint                    `json:"accepted"`
	Skipped  uint                    `json:"skipped"`
}

func ToWishlistProductResponse(item *models.WishlistItem) WishlistProductResponse {
	product := item.Product
	isAvailable := true
	availabilityReason := ""
	if product.DeletedAt != nil ||
		product.Category.DeletedAt != nil ||
		product.Brand.DeletedAt != nil ||
		!product.IsActive ||
		!product.IsListed ||
		product.ID <= 0 ||
		product.Category.ID <= 0 ||
		product.Brand.ID <= 0 ||
		!product.Category.IsActive ||
		!product.Brand.IsActive {
		isAvailable = false
		availabilityReason = "product_unavailable"
	}
	productName := product.Name
	if product.ID == 0 {
		productName = "Unavailable product"
	}
	var shortDescription, description, image string
	if product.ShortDescription != nil {
		shortDescription = *product.ShortDescription
	}
	if product.Description != nil {
		description = *product.Description
	}
	for _, img := range product.Images {
		if img.IsPrimary {
			image = img.ImageURL
			break
		}
	}
	return WishlistProductResponse{
		ID:                 item.ID,
		ProductID:          item.ProductID,
		ProductName:        productName,
		ProductSlug:        product.Slug,
		ShortDescription:   shortDescription,
		Description:        description,
		ImageURL:           image,
		IsAvailable:        isAvailable,
		AvailabilityReason: availabilityReason,
		AddedAt:            item.CreatedAt,
	}
}
func ToWishlistDetailsResponse(wishlist *models.Wishlist) WishlistDetailsResponse {
	result := WishlistDetailsResponse{
		Items: make([]WishlistProductResponse, 0),
	}
	if wishlist == nil {
		return result
	}
	result.ID = wishlist.ID
	result.Name = wishlist.Name
	result.CreatedAt = wishlist.CreatedAt
	result.UpdatedAt = wishlist.UpdatedAt
	for _, item := range wishlist.Items {
		result.Items = append(
			result.Items,
			ToWishlistProductResponse(&item),
		)
	}
	return result
}
func ToWishlistResponse(wishlist *models.Wishlist) WishlistResponse {
	return WishlistResponse{
		ID:        wishlist.ID,
		Name:      wishlist.Name,
		ItemCount: wishlist.ItemCount,
		CreatedAt: wishlist.CreatedAt,
		UpdatedAt: wishlist.UpdatedAt,
	}
}
