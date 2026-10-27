package response

import (
	"time"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type CouponUsageResponse struct {
	ID             uint      `json:"id"`
	CouponID       uuid.UUID `json:"coupon_id"`
	UserID         uint      `json:"user_id"`
	OrderID        uuid.UUID `json:"order_id"`
	DiscountAmount float64   `json:"discount_amount"`
	UsedAt         time.Time `json:"used_at"`
}

type CouponUsageListResponse struct {
	Usages     []CouponUsageResponse `json:"usages"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"total_pages"`
}

func ToCouponUsageResponse(
	usage *models.CouponUsage,
) CouponUsageResponse {
	return CouponUsageResponse{
		ID:             usage.ID,
		CouponID:       usage.CouponID,
		UserID:         usage.UserID,
		OrderID:        usage.OrderID,
		DiscountAmount: usage.DiscountAmount,
		UsedAt:         usage.UsedAt,
	}
}

func ToCouponUsageResponses(
	usages []models.CouponUsage,
) []CouponUsageResponse {
	responses := make([]CouponUsageResponse, 0, len(usages))

	for i := range usages {
		responses = append(
			responses,
			ToCouponUsageResponse(&usages[i]),
		)
	}

	return responses
}
