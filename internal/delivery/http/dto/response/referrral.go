package response

import "time"

type ReferralDashboardResponse struct {
	ReferralCode string             `json:"referral_code"`
	ReferralLink string             `json:"referral_link"`
	Summary      ReferralSummaryDTO `json:"summary"`
	Referrals    []ReferralItemDTO  `json:"referrals"`
	Pagination   PaginationDTO      `json:"pagination"`
}

type ReferralSummaryDTO struct {
	TotalReferrals     int64   `json:"total_referrals"`
	PendingReferrals   int64   `json:"pending_referrals"`
	CompletedReferrals int64   `json:"completed_referrals"`
	TotalRewardsEarned float64 `json:"total_rewards_earned"`
}

type ReferralItemDTO struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	RewardStatus string     `json:"reward_status"`
	RewardEarned float64    `json:"reward_earned"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type PaginationDTO struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
type AdminReferralDashboardResponse struct {
	TotalReferrals      int64   `json:"total_referrals"`
	PendingReferrals    int64   `json:"pending_referrals"`
	CompletedReferrals  int64   `json:"completed_referrals"`
	ReferrerRewardsPaid float64 `json:"referrer_rewards_paid"`
	ReferredRewardsPaid float64 `json:"referred_rewards_paid"`
	TotalRewardsPaid    float64 `json:"total_rewards_paid"`
	ConversionRate      float64 `json:"conversion_rate"`
}
type TopReferrersResponse struct {
	Referrers []TopReferrerDTO `json:"referrers"`
}

type TopReferrerDTO struct {
	ReferrerID          uint    `json:"referrer_id"`
	Name                string  `json:"name"`
	SuccessfulReferrals int64   `json:"successful_referrals"`
	RewardsEarned       float64 `json:"rewards_earned"`
}
type AdminReferralItemDTO struct {
	ID                   uint       `json:"id"`
	ReferrerID           uint       `json:"referrer_id"`
	ReferrerName         string     `json:"referrer_name"`
	ReferredUserID       uint       `json:"referred_user_id"`
	ReferredUserName     string     `json:"referred_user_name"`
	Status               string     `json:"status"`
	ReferrerReward       float64    `json:"referrer_reward"`
	ReferredReward       float64    `json:"referred_reward"`
	ReferrerRewardStatus string     `json:"referrer_reward_status"`
	ReferredRewardStatus string     `json:"referred_reward_status"`
	CreatedAt            time.Time  `json:"created_at"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
}

type AdminReferralsResponse struct {
	Referrals  []AdminReferralItemDTO `json:"referrals"`
	Pagination PaginationDTO          `json:"pagination"`
}
type ReferralSettingsResponse struct {
	ReferrerReward     float64   `json:"referrer_reward"`
	ReferredReward     float64   `json:"referred_reward"`
	MinimumOrderAmount float64   `json:"minimum_order_amount"`
	IsActive           bool      `json:"is_active"`
	UpdatedAt          time.Time `json:"updated_at"`
}
