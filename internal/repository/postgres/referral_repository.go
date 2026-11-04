package postgres

import (
	"context"
	"errors"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type referralRepository struct {
	db *gorm.DB
}

func NewReferralRepository(db *gorm.DB) *referralRepository {
	return &referralRepository{
		db: db,
	}
}

func (r *referralRepository) CreateIfNotReferred(ctx context.Context, referral *models.Referral) (bool, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "referred_user_id"},
			},
			DoNothing: true,
		}).
		Create(referral)

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}
func (r *referralRepository) GetByID(ctx context.Context, referralID uint) (*models.Referral, error) {
	var referral models.Referral
	err := r.db.WithContext(ctx).
		First(&referral, referralID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &referral, nil
}
func (r *referralRepository) GetByIDForUpdate(ctx context.Context, referralID uint) (*models.Referral, error) {
	var referral models.Referral
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&referral, referralID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &referral, nil
}
func (r *referralRepository) Update(ctx context.Context, referral *models.Referral) error {
	result := r.db.WithContext(ctx).
		Model(&models.Referral{}).
		Where("id = ?", referral.ID).
		Select(
			"Status",
			"ReferrerRewardStatus",
			"ReferredRewardStatus",
			"CompletedAt",
		).
		Updates(referral)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *referralRepository) GetByReferrerID(ctx context.Context, referrerID uint, status *models.ReferralStatus, rewardStatus *models.ReferralRewardStatus, limit int, offset int) ([]models.Referral, int64, error) {
	var referrals []models.Referral
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.Referral{}).
		Where("referrer_id = ?", referrerID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if rewardStatus != nil {
		query = query.Where("referrer_reward_status = ?", *rewardStatus)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Preload("ReferredUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&referrals).Error
	if err != nil {
		return nil, 0, err
	}
	return referrals, total, nil
}
func (r *referralRepository) GetReferralCounts(ctx context.Context, referrerID uint) (total int64, pending int64, completed int64, err error) {

	type ReferralCounts struct {
		Total     int64
		Pending   int64
		Completed int64
	}

	var counts ReferralCounts

	err = r.db.WithContext(ctx).
		Model(&models.Referral{}).
		Select(`
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = ?) AS pending,
			COUNT(*) FILTER (WHERE status = ?) AS completed
		`,
			models.ReferralStatusPending,
			models.ReferralStatusCompleted,
		).
		Where("referrer_id = ?", referrerID).
		Scan(&counts).Error

	if err != nil {
		return 0, 0, 0, err
	}

	return counts.Total, counts.Pending, counts.Completed, nil
}
func (r *referralRepository) GetTotalRewardsEarned(ctx context.Context, referrerID uint) (float64, error) {

	var total float64

	err := r.db.WithContext(ctx).
		Model(&models.Referral{}).
		Select("COALESCE(SUM(referrer_reward), 0)").
		Where("referrer_id = ?", referrerID).
		Where(
			"referrer_reward_status = ?",
			models.ReferralRewardStatusCompleted,
		).
		Scan(&total).Error

	if err != nil {
		return 0, err
	}

	return total, nil
}
func (r *referralRepository) GetReferralDashboardStats(ctx context.Context) (int64, int64, int64, float64, float64, error) {
	var stats struct {
		Total           int64
		Pending         int64
		Completed       int64
		ReferrerRewards float64
		ReferredRewards float64
	}

	err := r.db.WithContext(ctx).
		Model(&models.Referral{}).
		Select(`
            COUNT(*) AS total,
            COUNT(*) FILTER (WHERE status = ?) AS pending,
            COUNT(*) FILTER (WHERE status = ?) AS completed,
            COALESCE(SUM(referrer_reward) FILTER (
                WHERE referrer_reward_status = ?
            ), 0) AS referrer_rewards,
            COALESCE(SUM(referred_reward) FILTER (
                WHERE referred_reward_status = ?
            ), 0) AS referred_rewards
        `,
			models.ReferralStatusPending,
			models.ReferralStatusCompleted,
			models.ReferralRewardStatusCompleted,
			models.ReferralRewardStatusCompleted,
		).
		Scan(&stats).Error

	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	return stats.Total,
		stats.Pending,
		stats.Completed,
		stats.ReferrerRewards,
		stats.ReferredRewards,
		nil
}
func (r *referralRepository) GetTopReferrers(ctx context.Context, limit int) ([]models.TopReferrer, error) {
	if limit <= 0 {
		limit = 10
	}
	var referrers []models.TopReferrer
	err := r.db.WithContext(ctx).
		Model(&models.Referral{}).
		Select(`
            users.id AS referrer_id,
            users.name AS name,
            COUNT(referrals.id) FILTER (
                WHERE referrals.status = ?
            ) AS successful_referrals,
            COALESCE(SUM(referrals.referrer_reward) FILTER (
                WHERE referrals.referrer_reward_status = ?
            ), 0) AS rewards_earned
        `,
			models.ReferralStatusCompleted,
			models.ReferralRewardStatusCompleted,
		).
		Joins("JOIN users ON users.id = referrals.referrer_id").
		Group("users.id, users.name").
		Having(
			"COUNT(referrals.id) FILTER (WHERE referrals.status = ?) > 0",
			models.ReferralStatusCompleted,
		).
		Order("successful_referrals DESC").
		Limit(limit).
		Scan(&referrers).Error

	if err != nil {
		return nil, err
	}
	return referrers, nil
}
func (r *referralRepository) GetReferrals(ctx context.Context, status *models.ReferralStatus, rewardStatus *models.ReferralRewardStatus, referrerID *uint, startDate *time.Time, endDate *time.Time, limit int, offset int) ([]models.Referral, int64, error) {
	var referrals []models.Referral
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.Referral{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if rewardStatus != nil {
		query = query.Where(
			"(referrer_reward_status = ? OR referred_reward_status = ?)",
			*rewardStatus,
			*rewardStatus,
		)
	}
	if referrerID != nil {
		query = query.Where("referrer_id = ?", *referrerID)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	err := query.
		Preload("Referrer", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("ReferredUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&referrals).Error
	if err != nil {
		return nil, 0, err
	}
	return referrals, total, nil
}
func (r *referralRepository) GetPendingReferrals(ctx context.Context, limit int) ([]models.Referral, error) {
	var referrals []models.Referral
	err := r.db.WithContext(ctx).
		Where("referrals.status = ?", models.ReferralStatusPending).
		Where(`
			EXISTS (
				SELECT 1
				FROM orders
				WHERE orders.user_id = referrals.referred_user_id
				AND orders.order_status = ?
				AND orders.total_amount >= referrals.minimum_order_amount
			)
		`, "delivered").
		Order("referrals.created_at ASC").
		Limit(limit).
		Find(&referrals).Error
	return referrals, err
}

var _ interfaces.ReferralRepository = (*referralRepository)(nil)
