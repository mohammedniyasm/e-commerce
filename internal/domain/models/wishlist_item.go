package models

import "time"

type WishlistItem struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	WishlistID uint      `gorm:"not null;uniqueIndex:wishlist_items_index_0,priority:1"`
	ProductID  uint      `gorm:"not null;uniqueIndex:wishlist_items_index_0,priority:2"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`

	Wishlist Wishlist `gorm:"foreignKey:WishlistID"`
	Product  Product  `gorm:"foreignKey:ProductID"`
}
