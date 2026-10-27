package postgres

import (
	"context"
	"encoding/json"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(notification).Error
}
func (r *NotificationRepository) CreateRecipients(ctx context.Context, recipients []models.NotificationRecipient) error {
	if len(recipients) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&recipients).Error
}
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).
		First(&notification, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &notification, nil
}
func (r *NotificationRepository) ListByAdminID(ctx context.Context, adminID uint, limit int, page int, unreadOnly bool) ([]models.NotificationRecipient, int64, error) {
	var recipients []models.NotificationRecipient
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.NotificationRecipient{}).
		Where("admin_id = ?", adminID).
		Preload("Notification")

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	err := query.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&recipients).Error

	if err != nil {
		return nil, 0, err
	}

	return recipients, total, nil
}
func (r *NotificationRepository) CountUnreadByAdminID(ctx context.Context, adminID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.NotificationRecipient{}).
		Where("admin_id = ? AND is_read = ?", adminID, false).
		Count(&count).Error
	return count, err
}
func (r *NotificationRepository) MarkAsRead(ctx context.Context, adminID uint, notificationID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.NotificationRecipient{}).
		Where(
			"admin_id = ? AND notification_id = ?",
			adminID,
			notificationID,
		).
		Where("is_read = ?", false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	var recipient models.NotificationRecipient
	err := r.db.WithContext(ctx).
		Where(
			"admin_id = ? AND notification_id = ?",
			adminID,
			notificationID,
		).
		First(&recipient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainerrors.ErrRecordNotFound
		}
		return err
	}
	return nil
}
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, adminID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.NotificationRecipient{}).
		Where("admin_id = ? AND is_read = ?", adminID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}
func (r *NotificationRepository) ExistsByEventData(ctx context.Context, notificationType string, entityType string, entityID string, eventData json.RawMessage) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where(
			"type = ? AND entity_type = ? AND entity_id = ?",
			notificationType,
			entityType,
			entityID,
		).
		Where("event_data = ?::jsonb", string(eventData)).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

var _ interfaces.NotificationRepository = (*NotificationRepository)(nil)
