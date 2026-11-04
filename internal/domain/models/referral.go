package models

import "time"

type ReferralStatus string

const (
	ReferralStatusPending   ReferralStatus = "pending"
	ReferralStatusCompleted ReferralStatus = "completed"
)

type ReferralRewardStatus string

const (
	ReferralRewardStatusPending   ReferralRewardStatus = "pending"
	ReferralRewardStatusCompleted ReferralRewardStatus = "completed"
)

type TopReferrer struct {
	ReferrerID          uint
	Name                string
	SuccessfulReferrals int64
	RewardsEarned       float64
}
type Referral struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	ReferrerID     uint `gorm:"not null;index"`
	ReferredUserID uint `gorm:"not null;uniqueIndex"`

	Status ReferralStatus `gorm:"type:varchar(20);not null;default:pending"`

	ReferrerReward     float64 `gorm:"type:numeric(12,2);not null"`
	ReferredReward     float64 `gorm:"type:numeric(12,2);not null"`
	MinimumOrderAmount float64 `gorm:"type:numeric(12,2);not null"`

	ReferrerRewardStatus ReferralRewardStatus `gorm:"type:varchar(20);not null;default:pending"`
	ReferredRewardStatus ReferralRewardStatus `gorm:"type:varchar(20);not null;default:pending"`
	CreatedAt            time.Time            `gorm:"not null;autoCreateTime"`
	CompletedAt          *time.Time

	Referrer     User `gorm:"foreignKey:ReferrerID;references:ID"`
	ReferredUser User `gorm:"foreignKey:ReferredUserID;references:ID"`
}
