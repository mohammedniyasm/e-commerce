package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatusHistory struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Status    string    `gorm:"type:order_status;not null"`
	Note      *string   `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`

	Order Order `gorm:"foreignKey:OrderID"`
}

func (OrderStatusHistory) TableName() string {
	return "order_status_history"
}
