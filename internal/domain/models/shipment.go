package models

import (
	"time"

	"github.com/google/uuid"
)

type Shipment struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	OrderID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	CourierName    *string   `gorm:"size:255"`
	TrackingNumber *string   `gorm:"size:255;index"`
	ShippingStatus string    `gorm:"type:shipping_status;not null;default:pending"`
	ShippedAt      *time.Time
	DeliveredAt    *time.Time
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`

	Order   Order                     `gorm:"foreignKey:OrderID"`
	History []ShipmentTrackingHistory `gorm:"foreignKey:ShipmentID"`
}
