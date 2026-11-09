package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID              uint            `gorm:"primaryKey;autoIncrement"`
	OrderID         uuid.UUID       `gorm:"type:uuid;not null;index"`
	PaymentMethod   string          `gorm:"not null;size:30"`
	PaymentStatus   string          `gorm:"type:payment_status;not null"`
	GatewayOrderID  *string         `gorm:"size:255;index"`
	TransactionID   *string         `gorm:"size:255;uniqueIndex"`
	GatewayResponse json.RawMessage `gorm:"type:jsonb"`
	FailureReason   *string         `gorm:"size:255"`
	Amount          float64         `gorm:"type:numeric(12,2);not null"`
	PaidAt          *time.Time
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	Order Order `gorm:"foreignKey:OrderID"`
}
