package models

import "time"

type CartItem struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	CartID    uint      `gorm:"not null;uniqueIndex:cart_items_index_0,priority:1"`
	VariantID uint      `gorm:"not null;uniqueIndex:cart_items_index_0,priority:2"`
	Quantity  int       `gorm:"not null;check:chk_cart_items_quantity,quantity > 0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Cart    Cart           `gorm:"foreignKey:CartID"`
	Variant ProductVariant `gorm:"foreignKey:VariantID"`
}
