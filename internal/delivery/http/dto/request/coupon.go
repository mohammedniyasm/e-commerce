package request

type CreateCouponRequest struct {
	Code               string   `json:"code" binding:"required,max=50"`
	Description        *string  `json:"description"`
	DiscountType       string   `json:"discount_type" binding:"required,oneof=percentage fixed"`
	DiscountValue      float64  `json:"discount_value" binding:"required,gt=0"`
	MinimumOrderAmount float64  `json:"minimum_order_amount" binding:"gte=0"`
	MaximumDiscount    *float64 `json:"maximum_discount" binding:"omitempty,gt=0"`
	UsageLimit         *int     `json:"usage_limit" binding:"omitempty,gt=0"`
	UsageLimitPerUser  *int     `json:"usage_limit_per_user" binding:"omitempty,gt=0"`
	StartAt            string   `json:"start_at" binding:"required"`
	ExpiresAt          string   `json:"expires_at" binding:"required"`
	IsActive           bool     `json:"is_active"`
}
type UpdateCouponRequest struct {
	Code               string   `json:"code" binding:"required,max=50"`
	Description        *string  `json:"description"`
	DiscountType       string   `json:"discount_type" binding:"required,oneof=percentage fixed"`
	DiscountValue      float64  `json:"discount_value" binding:"required,gt=0"`
	MinimumOrderAmount float64  `json:"minimum_order_amount" binding:"gte=0"`
	MaximumDiscount    *float64 `json:"maximum_discount" binding:"omitempty,gt=0"`
	UsageLimit         *int     `json:"usage_limit" binding:"omitempty,gt=0"`
	UsageLimitPerUser  *int     `json:"usage_limit_per_user" binding:"omitempty,gt=0"`
	StartAt            string   `json:"start_at" binding:"required"`
	ExpiresAt          string   `json:"expires_at" binding:"required"`
	IsActive           bool     `json:"is_active"`
}
type ValidateCouponRequest struct {
	Code        string  `json:"code" binding:"required,max=50"`
	OrderAmount float64 `json:"order_amount" binding:"gte=0"`
}
type ListCouponRequest struct {
	Search     string `form:"search"`
	ActiveOnly *bool  `form:"active_only"`
	Limit      int    `form:"limit"`
	Page       int    `form:"page"`
	Sort       string `form:"sort"`
}
