package models

import "time"

type WalletTopUpStatus string

const (
	WalletTopUpPending WalletTopUpStatus = "pending"
	WalletTopUpPaid    WalletTopUpStatus = "paid"
	WalletTopUpFailed  WalletTopUpStatus = "failed"
)

type WalletTopUp struct {
	ID               uint              `gorm:"primaryKey;autoIncrement"`
	WalletID         uint              `gorm:"not null;index"`
	Amount           float64           `gorm:"type:numeric(12,2);not null"`
	GatewayOrderID   string            `gorm:"size:255;not null;uniqueIndex"`
	GatewayPaymentID *string           `gorm:"size:255;uniqueIndex"`
	Status           WalletTopUpStatus `gorm:"type:varchar(20);not null;default:pending"`
	PaidAt           *time.Time
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`

	Wallet Wallet `gorm:"foreignKey:WalletID;references:ID"`
}
