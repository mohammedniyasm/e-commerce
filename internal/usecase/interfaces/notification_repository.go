package interfaces

import (
	"context"
	"encoding/json"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	// Notification creation and retrieval
	Create(ctx context.Context, notification *models.Notification) error
	CreateRecipients(ctx context.Context, recipients []models.NotificationRecipient) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	// Admin-specific notification listing
	ListByAdminID(ctx context.Context, adminID uint, limit int, page int, unreadOnly bool) ([]models.NotificationRecipient, int64, error)
	CountUnreadByAdminID(ctx context.Context, adminID uint) (int64, error)
	// Admin-specific read status
	MarkAsRead(ctx context.Context, adminID uint, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, adminID uint) error
	ExistsByEventData(ctx context.Context, notificationType string, entityType string, entityID string, eventData json.RawMessage) (bool, error)
}
