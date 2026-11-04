package models

import "time"

type Wallet struct {
	ID        uint      `gorm:"primaryKey"`
	Balance   float64   `gorm:"type:numeric(12,2);not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	User         User                `gorm:"foreignKey:ID;references:ID"`
	Transactions []WalletTransaction `gorm:"foreignKey:WalletID;references:ID"`
}
