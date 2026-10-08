package models

import (
	"time"
)

type Brand struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	Name        string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description *string    `gorm:"type:text"`
	Logo        *string    `gorm:"type:varchar(255)"`
	IsActive    bool       `gorm:"not null;default:true;index"`
	DeletedAt   *time.Time `gorm:"index"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
