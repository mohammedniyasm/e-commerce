package models

import (
	"time"
)

type ShipmentTrackingHistory struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	ShipmentID     uint      `gorm:"not null;index"`
	ShippingStatus string    `gorm:"type:shipping_status;not null"`
	Description    *string   `gorm:"type:text"`
	Location       *string   `gorm:"size:255"`
	CreatedAt      time.Time `gorm:"not null"`

	Shipment Shipment `gorm:"foreignKey:ShipmentID"`
}

func (ShipmentTrackingHistory) TableName() string {
	return "shipment_tracking_history"
}
