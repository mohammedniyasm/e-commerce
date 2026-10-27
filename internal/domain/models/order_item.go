package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID          uint            `gorm:"primaryKey;autoIncrement"`
	OrderID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	VariantID   uint            `gorm:"not null"`
	ProductName string          `gorm:"not null;size:255"`
	ImageURL    string          `gorm:"not null;size:500"`
	VariantInfo json.RawMessage `gorm:"type:jsonb;not null"`
	UnitPrice   float64         `gorm:"type:numeric(12,2);not null"`
	Quantity    int             `gorm:"not null"`
	Subtotal    float64         `gorm:"type:numeric(12,2);not null"`
	CreatedAt   time.Time       `gorm:"not null"`
	UpdatedAt   time.Time       `gorm:"not null"`

	Order   Order          `gorm:"foreignKey:OrderID"`
	Variant ProductVariant `gorm:"foreignKey:VariantID"`
}
