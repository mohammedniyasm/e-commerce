package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderAddress struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	OrderID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Name         string    `gorm:"not null;size:255"`
	Phone        string    `gorm:"not null;size:20"`
	AddressLine1 string    `gorm:"not null;size:255"`
	AddressLine2 *string   `gorm:"size:255"`
	City         string    `gorm:"not null;size:255"`
	State        string    `gorm:"not null;size:255"`
	PostalCode   string    `gorm:"not null;size:20"`
	Country      string    `gorm:"not null;size:100"`
	CreatedAt    time.Time `gorm:"not null"`

	Order Order `gorm:"foreignKey:OrderID"`
}
