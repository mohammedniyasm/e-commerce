package models

import "time"

type Cart struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	User  User       `gorm:"foreignKey:UserID"`
	Items []CartItem `gorm:"foreignKey:CartID"`
}

func (Cart) TableName() string {
	return "cart"
}
