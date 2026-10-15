package response

import (
	"time"

	"ecommerce/internal/domain/models"
)

type CategoryResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"total_pages"`
}

func ToCategoryResponse(category *models.Category) CategoryResponse {
	return CategoryResponse{
		ID:          uint(category.ID),
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}

type StoreCategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func ToStoreCategoryResponse(
	category models.Category,
) StoreCategoryResponse {
	return StoreCategoryResponse{
		ID:   uint(category.ID),
		Name: category.Name,
	}
}

func ToStoreCategoriesResponse(
	categories []models.Category,
) []StoreCategoryResponse {
	responses := make([]StoreCategoryResponse, 0, len(categories))

	for _, category := range categories {
		responses = append(
			responses,
			ToStoreCategoryResponse(category),
		)
	}

	return responses
}
