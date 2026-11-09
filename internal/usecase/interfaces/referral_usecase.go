package interfaces

import (
	"context"
	"time"

	"ecommerce/internal/domain/models"
)

type ReferralUseCase interface {
	GetReferralDashboard(
		ctx context.Context,
		userID uint,
		status *models.ReferralStatus,
		rewardStatus *models.ReferralRewardStatus,
		limit int,
		offset int,
	) (
		referralCode string,
		total int64,
		pending int64,
		completed int64,
		totalRewards float64,
		referrals []models.Referral,
		filteredTotal int64,
		err error,
	)
	GetAdminReferralDashboard(ctx context.Context) (total int64, pending int64, completed int64, referrerRewards float64, referredRewards float64, totalRewardsPaid float64, conversionRate float64, err error)
	GetTopReferrers(ctx context.Context, limit int) ([]models.TopReferrer, error)
	GetReferrals(ctx context.Context, status *models.ReferralStatus, rewardStatus *models.ReferralRewardStatus, referrerID *uint, startDate *time.Time, endDate *time.Time, limit int, offset int) ([]models.Referral, int64, error)
	GetReferralSettings(ctx context.Context) (*models.ReferralSettings, error)
	UpdateReferralSettings(ctx context.Context, referrerReward float64, referredReward float64, minimumOrderAmount float64, isActive bool) (*models.ReferralSettings, error)
}
