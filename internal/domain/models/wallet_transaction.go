package models

import (
	"time"

	"github.com/google/uuid"
)

type WalletTransactionType string

const (
	WalletTransactionCredit WalletTransactionType = "credit"
	WalletTransactionDebit  WalletTransactionType = "debit"
)

type WalletTransactionSource string

const (
	WalletSourceOrderPayment    WalletTransactionSource = "order_payment"
	WalletSourceRefund          WalletTransactionSource = "refund"
	WalletSourceReferralReward  WalletTransactionSource = "referral_reward"
	WalletSourceTopUp           WalletTransactionSource = "top_up"
	WalletSourceAdminAdjustment WalletTransactionSource = "admin_adjustment"
)

type WalletTransaction struct {
	ID              uint                    `gorm:"primaryKey;autoIncrement"`
	WalletID        uint                    `gorm:"not null;index"`
	TransactionType WalletTransactionType   `gorm:"type:varchar(20);not null"`
	Source          WalletTransactionSource `gorm:"type:varchar(30);not null"`
	Amount          float64                 `gorm:"type:numeric(12,2);not null"`
	BalanceBefore   float64                 `gorm:"type:numeric(12,2);not null"`
	BalanceAfter    float64                 `gorm:"type:numeric(12,2);not null"`
	ReferenceID     *uuid.UUID              `gorm:"type:uuid;index"`
	ReferralID      *uint                   `gorm:"index"`
	ReturnRequestID *uint                   `gorm:"uniqueIndex:idx_wallet_transactions_return_request_id"`
	Description     *string                 `gorm:"type:text"`
	CreatedAt       time.Time               `gorm:"not null"`

	Wallet Wallet `gorm:"foreignKey:WalletID;references:ID"`
}
