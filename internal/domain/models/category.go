package models

import "time"

type Category struct {
	ID          int64      `gorm:"primaryKey;autoIncrement"`
	Name        string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description *string    `gorm:"type:text"`
	IsActive    bool       `gorm:"not null;default:true"`
	DeletedAt   *time.Time `gorm:"index"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
}
