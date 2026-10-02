package models

import "time"

type Address struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   `gorm:"not null"`
	Name         string `gorm:"not null;size:255"`
	Phone        string `gorm:"not null;size:20"`
	AddressLine1 string `gorm:"not null;size:255"`
	AddressLine2 string `gorm:"column:address_line_2;size:255"`
	City         string `gorm:"not null;size:255"`
	PostalCode   string `gorm:"not null;size:20"`
	State        string `gorm:"not null;size:255"`
	Country      string `gorm:"not null;default:India;size:255"`
	IsDefault    bool   `gorm:"not null;default:false"`
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
