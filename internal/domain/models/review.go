package models

import "time"

type Review struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	ProductID   uint `gorm:"not null;index"`
	UserID      uint `gorm:"not null;index"`
	OrderItemID uint `gorm:"not null;uniqueIndex"`

	Rating  int    `gorm:"not null"`
	Comment string `gorm:"type:text;not null"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}
