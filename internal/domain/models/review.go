package models

import "time"

type Review struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	ProductID   uint      `gorm:"not null;index:reviews_index_0"`
	UserID      uint      `gorm:"not null;index:reviews_index_1"`
	OrderItemID uint      `gorm:"not null;uniqueIndex:uq_reviews_order_item"`
	Rating      int16     `gorm:"type:smallint;not null"`
	Comment     *string   `gorm:"type:text"`
	IsVisible   bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"not null;autoUpdateTime"`

	Product   Product   `gorm:"foreignKey:ProductID"`
	User      User      `gorm:"foreignKey:UserID"`
	OrderItem OrderItem `gorm:"foreignKey:OrderItemID"`
}
