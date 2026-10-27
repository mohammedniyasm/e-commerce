package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationRecipient struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	NotificationID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_notification_recipient"`
	AdminID        uint      `gorm:"not null;uniqueIndex:idx_notification_recipient;index"`
	IsRead         bool      `gorm:"not null;default:false"`
	ReadAt         *time.Time

	Notification Notification `gorm:"foreignKey:NotificationID"`
}
