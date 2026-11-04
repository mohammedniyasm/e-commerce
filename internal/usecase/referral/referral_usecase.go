package referral

import (
	"context"
	"log/slog"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
)

type referralUseCase struct {
	userRepo     interfaces.UserRepository
	referralRepo interfaces.ReferralRepository
	settingsRepo interfaces.ReferralSettingsRepository
	log          *slog.Logger
}

func NewReferralUseCase(
	userRepo interfaces.UserRepository,
	referralRepo interfaces.ReferralRepository,
	settingsRepo interfaces.ReferralSettingsRepository,
	log *slog.Logger,
) interfaces.ReferralUseCase {
	return &referralUseCase{
		userRepo:     userRepo,
		referralRepo: referralRepo,
		settingsRepo: settingsRepo,
		log:          log,
	}
}
func (u *referralUseCase) GetReferralDashboard(ctx context.Context, userID uint, status *models.ReferralStatus, rewardStatus *models.ReferralRewardStatus, limit int, offset int) (string, int64, int64, int64, float64, []models.Referral, int64, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", 0, 0, 0, 0, nil, 0, err
	}
	total, pending, completed, err :=
		u.referralRepo.GetReferralCounts(ctx, userID)
	if err != nil {
		return "", 0, 0, 0, 0, nil, 0, err
	}
	totalRewards, err :=
		u.referralRepo.GetTotalRewardsEarned(ctx, userID)
	if err != nil {
		return "", 0, 0, 0, 0, nil, 0, err
	}
	referrals, filteredTotal, err :=
		u.referralRepo.GetByReferrerID(
			ctx,
			userID,
			status,
			rewardStatus,
			limit,
			offset,
		)
	if err != nil {
		return "", 0, 0, 0, 0, nil, 0, err
	}
	return user.ReferralCode,
		total,
		pending,
		completed,
		totalRewards,
		referrals,
		filteredTotal,
		nil
}
func (u *referralUseCase) GetAdminReferralDashboard(ctx context.Context) (int64, int64, int64, float64, float64, float64, float64, error) {
	total, pending, completed, referrerRewards, referredRewards, err :=
		u.referralRepo.GetReferralDashboardStats(ctx)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, err
	}
	totalRewardsPaid := referrerRewards + referredRewards
	var conversionRate float64
	if total > 0 {
		conversionRate = float64(completed) / float64(total) * 100
	}
	return total, pending, completed,
		referrerRewards, referredRewards,
		totalRewardsPaid, conversionRate, nil
}
func (u *referralUseCase) GetTopReferrers(ctx context.Context, limit int) ([]models.TopReferrer, error) {
	return u.referralRepo.GetTopReferrers(ctx, limit)
}
func (u *referralUseCase) GetReferrals(ctx context.Context, status *models.ReferralStatus, rewardStatus *models.ReferralRewardStatus, referrerID *uint, startDate *time.Time, endDate *time.Time, limit int, offset int) ([]models.Referral, int64, error) {
	return u.referralRepo.GetReferrals(
		ctx,
		status,
		rewardStatus,
		referrerID,
		startDate,
		endDate,
		limit,
		offset,
	)
}
func (u *referralUseCase) GetReferralSettings(ctx context.Context) (*models.ReferralSettings, error) {
	return u.settingsRepo.Get(ctx)
}
func (u *referralUseCase) UpdateReferralSettings(ctx context.Context, referrerReward float64, referredReward float64, minimumOrderAmount float64, isActive bool) (*models.ReferralSettings, error) {
	if referrerReward < 0 || referredReward < 0 || minimumOrderAmount < 0 {
		return nil, domainerrors.ErrInvalidReferralSettings
	}
	settings, err := u.settingsRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	settings.ReferrerReward = referrerReward
	settings.ReferredReward = referredReward
	settings.MinimumOrderAmount = minimumOrderAmount
	settings.IsActive = isActive
	if err := u.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}
	return settings, nil
}
