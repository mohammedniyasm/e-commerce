package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type NotificationUseCase interface {
	// Notification creation
	CreateNotification(ctx context.Context, notification *models.Notification) error
	// Notification retrieval
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	ListByAdminID(ctx context.Context, adminID uint, limit int, page int, unreadOnly bool) ([]models.NotificationRecipient, int64, error)
	CountUnreadByAdminID(ctx context.Context, adminID uint) (int64, error)
	// Read status management
	MarkAsRead(ctx context.Context, adminID uint, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, adminID uint) error
}
