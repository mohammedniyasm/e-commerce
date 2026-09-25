package models

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Phone     string `gorm:"size:20"`
	Password  *string
	Role      string `gorm:"not null;default:user"`
	IsBlocked bool   `gorm:"not null;default:false"`
	LastSeen  *time.Time
	CreatedAt time.Time
	UpdatedAt *time.Time
}
