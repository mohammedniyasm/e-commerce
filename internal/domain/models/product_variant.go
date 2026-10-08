package models

import "time"

type ProductVariant struct {
	ID        uint `gorm:"primaryKey"`
	ProductID uint `gorm:"not null;index"`

	Size  *string `gorm:"size:50"`
	Color *string `gorm:"size:100"`

	SKU string `gorm:"not null;size:100;uniqueIndex"`

	MRP          float64 `gorm:"not null"`
	SellingPrice float64 `gorm:"not null"`
	Stock        int     `gorm:"not null;default:0"`

	IsActive bool `gorm:"not null;default:true;index"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`

	Images []ProductImage `gorm:"foreignKey:VariantID"`
}
