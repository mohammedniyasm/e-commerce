package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
	"time"
)

type ReferralRepository interface {
	CreateIfNotReferred(ctx context.Context, referral *models.Referral) (bool, error)
	GetByID(ctx context.Context, referralID uint) (*models.Referral, error)
	GetByIDForUpdate(ctx context.Context, referralID uint) (*models.Referral, error)
	Update(ctx context.Context, referral *models.Referral) error
	GetByReferrerID(ctx context.Context, referrerID uint, status *models.ReferralStatus, rewardStatus *models.ReferralRewardStatus, limit int, offset int) ([]models.Referral, int64, error)
	GetReferralCounts(ctx context.Context, referrerID uint) (total int64, pending int64, completed int64, err error)
	GetTotalRewardsEarned(ctx context.Context, referrerID uint) (float64, error)
	GetReferralDashboardStats(ctx context.Context) (total int64, pending int64, completed int64, referrerRewards float64, referredRewards float64, err error)
	GetTopReferrers(ctx context.Context, limit int) ([]models.TopReferrer, error)
	GetReferrals(ctx context.Context, status *models.ReferralStatus, rewardStatus *models.ReferralRewardStatus, referrerID *uint, startDate *time.Time, endDate *time.Time, limit int, offset int) ([]models.Referral, int64, error)
	GetPendingReferrals(ctx context.Context, limit int) ([]models.Referral, error)
}
