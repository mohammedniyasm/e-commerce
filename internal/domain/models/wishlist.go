package models

import "time"

type Wishlist struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null;index"`
	Name      string    `gorm:"type:varchar(100);not null"`
	ItemCount int64     `gorm:"-"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	User  User           `gorm:"foreignKey:UserID"`
	Items []WishlistItem `gorm:"foreignKey:WishlistID"`
}

func (Wishlist) TableName() string {
	return "wishlist"
}
