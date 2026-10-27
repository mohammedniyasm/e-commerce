package models

import (
	"time"

	"github.com/google/uuid"
)

type CouponUsage struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	CouponID       uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID         uint      `gorm:"not null;index"`
	OrderID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	DiscountAmount float64   `gorm:"type:numeric(12,2);not null"`
	UsedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	Status         bool      `gorm:"not null;default:true"`

	Coupon Coupon `gorm:"foreignKey:CouponID"`
	Order  Order  `gorm:"foreignKey:OrderID"`
}
