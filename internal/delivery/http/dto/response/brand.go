package response

import (
	"time"

	"ecommerce/internal/domain/models"
)

type BrandResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Logo        *string   `json:"logo"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BrandListResponse struct {
	Brands     []BrandResponse `json:"brands"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	Total      int64           `json:"total"`
	TotalPages int             `json:"total_pages"`
}

func ToBrandResponse(brand *models.Brand) BrandResponse {
	return BrandResponse{
		ID:          uint(brand.ID),
		Name:        brand.Name,
		Description: brand.Description,
		Logo:        brand.Logo,
		IsActive:    brand.IsActive,
		CreatedAt:   brand.CreatedAt,
		UpdatedAt:   brand.UpdatedAt,
	}
}
