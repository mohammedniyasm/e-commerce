package models

import "time"

type ProductImage struct {
	ID uint `gorm:"primaryKey"`

	ProductID uint  `gorm:"not null;index"`
	VariantID *uint `gorm:"index"`

	ImageURL string `gorm:"not null;size:500"`

	DisplayOrder int  `gorm:"not null;default:0"`
	IsPrimary    bool `gorm:"not null;default:false"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`
	Variant *ProductVariant `gorm:"foreignKey:VariantID"`
}