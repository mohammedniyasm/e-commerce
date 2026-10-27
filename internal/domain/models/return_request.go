package models

import (
	"time"

	"github.com/google/uuid"
)

type ReturnRequest struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null;index"`
	OrderItemID uint      `gorm:"not null;index"`

	RequestType        string `gorm:"type:varchar(20);not null"`
	RequestedVariantID *uint  `gorm:"index"`

	Quantity int    `gorm:"not null;default:1;check:quantity > 0"`
	Reason   string `gorm:"type:text;not null"`
	Status   string `gorm:"type:varchar(30);not null;default:requested;index"`

	AdminNote *string `gorm:"type:text"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`

	Order            Order           `gorm:"foreignKey:OrderID"`
	OrderItem        OrderItem       `gorm:"foreignKey:OrderItemID"`
	RequestedVariant *ProductVariant `gorm:"foreignKey:RequestedVariantID"`
}
