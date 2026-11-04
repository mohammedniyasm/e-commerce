package response

import (
	"time"

	"ecommerce/internal/domain/models"
)

type ReviewResponse struct {
	ID          uint      `json:"id"`
	ProductID   uint      `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"`
	UserName    string    `json:"user_name,omitempty"`
	OrderItemID uint      `json:"order_item_id"`
	Rating      int16     `json:"rating"`
	Comment     *string   `json:"comment"`
	IsVisible   bool      `json:"is_visible"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ReviewListResponse struct {
	Reviews    []ReviewResponse `json:"reviews"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"total_pages"`
}

func ToReviewResponse(review *models.Review) ReviewResponse {
	result := ReviewResponse{
		ID:          review.ID,
		ProductID:   review.ProductID,
		OrderItemID: review.OrderItemID,
		Rating:      review.Rating,
		Comment:     review.Comment,
		IsVisible:   review.IsVisible,
		CreatedAt:   review.CreatedAt,
		UpdatedAt:   review.UpdatedAt,
	}
	return result
}

func ToReviewResponses(reviews []models.Review) []ReviewResponse {
	result := make([]ReviewResponse, 0, len(reviews))
	for i := range reviews {
		result = append(result, ToReviewResponse(&reviews[i]))
	}

	return result
}
