package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID uint `gorm:"primaryKey"`

	Name             string  `gorm:"not null;size:255"`
	Slug             string  `gorm:"not null;size:255;uniqueIndex"`
	ShortDescription *string `gorm:"type:text"`
	Description      *string `gorm:"type:text"`

	CategoryID uint `gorm:"not null;index"`
	BrandID    uint `gorm:"not null;index"`

	IsActive bool `gorm:"not null;default:true;index"`
	IsListed bool `gorm:"not null;default:true;index"`

	DeletedAt gorm.DeletedAt `gorm:"index"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Category Category `gorm:"foreignKey:CategoryID"`
	// Brand    Brand    `gorm:"foreignKey:BrandID"`

	Variants []ProductVariant `gorm:"foreignKey:ProductID"`
	Images   []ProductImage   `gorm:"foreignKey:ProductID"`
}
