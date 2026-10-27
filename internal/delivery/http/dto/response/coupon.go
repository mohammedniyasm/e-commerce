package response

import (
	"time"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type CouponResponse struct {
	ID                 uuid.UUID `json:"id"`
	Code               string    `json:"code"`
	Description        *string   `json:"description,omitempty"`
	DiscountType       string    `json:"discount_type"`
	DiscountValue      float64   `json:"discount_value"`
	MinimumOrderAmount float64   `json:"minimum_order_amount"`
	MaximumDiscount    *float64  `json:"maximum_discount,omitempty"`
	UsageLimit         *int      `json:"usage_limit,omitempty"`
	UsageLimitPerUser  *int      `json:"usage_limit_per_user,omitempty"`
	StartAt            time.Time `json:"start_at"`
	ExpiresAt          time.Time `json:"expires_at"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CouponListResponse struct {
	Coupons    []CouponResponse `json:"coupons"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"total_pages"`
}

type ValidateCouponResponse struct {
	CouponID       uuid.UUID `json:"coupon_id"`
	Code           string    `json:"code"`
	DiscountType   string    `json:"discount_type"`
	DiscountValue  float64   `json:"discount_value"`
	OrderAmount    float64   `json:"order_amount"`
	DiscountAmount float64   `json:"discount_amount"`
	FinalAmount    float64   `json:"final_amount"`
}

func ToCouponResponse(coupon *models.Coupon) CouponResponse {
	return CouponResponse{
		ID:                 coupon.ID,
		Code:               coupon.Code,
		Description:        coupon.Description,
		DiscountType:       coupon.DiscountType,
		DiscountValue:      coupon.DiscountValue,
		MinimumOrderAmount: coupon.MinimumOrderAmount,
		MaximumDiscount:    coupon.MaximumDiscount,
		UsageLimit:         coupon.UsageLimit,
		UsageLimitPerUser:  coupon.UsageLimitPerUser,
		StartAt:            coupon.StartAt,
		ExpiresAt:          coupon.ExpiresAt,
		IsActive:           coupon.IsActive,
		CreatedAt:          coupon.CreatedAt,
		UpdatedAt:          coupon.UpdatedAt,
	}
}

func ToCouponResponses(coupons []models.Coupon) []CouponResponse {
	responses := make([]CouponResponse, 0, len(coupons))

	for i := range coupons {
		responses = append(
			responses,
			ToCouponResponse(&coupons[i]),
		)
	}

	return responses
}
