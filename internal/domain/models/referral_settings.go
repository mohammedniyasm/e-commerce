package models

import "time"

type ReferralSettings struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement:false"`
	ReferrerReward     float64   `gorm:"type:numeric(12,2);not null;default:0"`
	ReferredReward     float64   `gorm:"type:numeric(12,2);not null;default:0"`
	MinimumOrderAmount float64   `gorm:"type:numeric(12,2);not null;default:0"`
	IsActive           bool      `gorm:"not null;default:false"`
	UpdatedAt          time.Time `gorm:"not null;autoUpdateTime"`
}
