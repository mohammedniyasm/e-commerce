package request

type UpdateReferralSettingsRequest struct {
	ReferrerReward     float64 `json:"referrer_reward" binding:"required,gte=0"`
	ReferredReward     float64 `json:"referred_reward" binding:"required,gte=0"`
	MinimumOrderAmount float64 `json:"minimum_order_amount" binding:"required,gte=0"`
	IsActive           *bool   `json:"is_active" binding:"required"`
}
