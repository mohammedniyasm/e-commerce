package models

import (
	"time"

	"github.com/google/uuid"
)

type Coupon struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code               string    `gorm:"not null;uniqueIndex;size:50"`
	Description        *string   `gorm:"type:text"`
	DiscountType       string    `gorm:"type:coupon_type;not null"`
	DiscountValue      float64   `gorm:"type:numeric(12,2);not null"`
	MinimumOrderAmount float64   `gorm:"type:numeric(12,2);not null"`
	MaximumDiscount    *float64  `gorm:"type:numeric(12,2)"`
	UsageLimit         *int      `gorm:""`
	UsageLimitPerUser  *int      `gorm:""`
	StartAt            time.Time `gorm:"not null"`
	ExpiresAt          time.Time `gorm:"not null"`
	IsActive           bool      `gorm:"not null;default:false"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
}
