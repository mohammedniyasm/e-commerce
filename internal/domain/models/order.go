package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uint       `gorm:"not null;index"`
	OrderNumber     string     `gorm:"not null;uniqueIndex;size:255"`
	Subtotal        float64    `gorm:"type:numeric(12,2);not null"`
	DiscountAmount  float64    `gorm:"type:numeric(12,2);not null;default:0"`
	ShippingAmount  float64    `gorm:"type:numeric(12,2);not null;default:0"`
	TotalAmount     float64    `gorm:"type:numeric(12,2);not null"`
	OrderStatus     string     `gorm:"type:order_status;not null"`
	PaymentStatus   string     `gorm:"type:payment_status;not null"`
	PaymentDeadline *time.Time `gorm:"column:payment_deadline"`

	PlacedAt  time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Items         []OrderItem          `gorm:"foreignKey:OrderID"`
	Address       *OrderAddress        `gorm:"foreignKey:OrderID"`
	Payments      []Payment            `gorm:"foreignKey:OrderID"`
	StatusHistory []OrderStatusHistory `gorm:"foreignKey:OrderID"`
	Shipment      *Shipment            `gorm:"foreignKey:OrderID"`
}
