package models

import (
	"time"

	"github.com/google/uuid"
)

type ShipmentType string

const (
	ShipmentTypeDelivery         ShipmentType = "delivery"
	ShipmentTypeReturn           ShipmentType = "return"
	ShipmentTypeReturnToCustomer ShipmentType = "return_to_customer"
)

type Shipment struct {
	ID              uint         `gorm:"primaryKey;autoIncrement"`
	OrderID         uuid.UUID    `gorm:"type:uuid;not null;Index"`
	ReturnRequestID *uint        `gorm:"index"`
	Type            ShipmentType `gorm:"type:varchar(20);not null;default:delivery"`
	CourierName     *string      `gorm:"size:255"`
	TrackingNumber  *string      `gorm:"size:255;index"`
	ShippingStatus  string       `gorm:"type:shipping_status;not null;default:pending"`
	ShippedAt       *time.Time
	DeliveredAt     *time.Time
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	ReturnRequest *ReturnRequest            `gorm:"foreignKey:ReturnRequestID;references:ID"`
	Order         Order                     `gorm:"foreignKey:OrderID"`
	History       []ShipmentTrackingHistory `gorm:"foreignKey:ShipmentID"`
}
