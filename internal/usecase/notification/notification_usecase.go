package notification

import (
	"context"
	"log/slog"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
)

type NotificationUseCase struct {
	notificationRepo interfaces.NotificationRepository
	unitOfWork       interfaces.UnitOfWork
	log              *slog.Logger
}

func NewNotificationUseCase(
	notificationRepo interfaces.NotificationRepository,
	unitOfWork interfaces.UnitOfWork,
	log *slog.Logger,
) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
		unitOfWork:       unitOfWork,
		log:              log,
	}
}
func (u *NotificationUseCase) CreateNotification(ctx context.Context, notification *models.Notification) error {
	if notification == nil {
		return domainerrors.ErrInvalidInput
	}
	return u.unitOfWork.WithinTransaction(
		ctx,
		func(repos interfaces.TransactionRepositories) error {
			adminIDs, err := repos.User.ListAdminIDs(ctx)
			if err != nil {
				u.log.ErrorContext(
					ctx,
					"failed to retrieve admin IDs for notification",
					"error", err,
				)
				return err
			}
			if notification.ID == uuid.Nil {
				notification.ID = uuid.New()
			}
			if err := repos.Notification.Create(ctx, notification); err != nil {
				u.log.ErrorContext(
					ctx,
					"failed to create notification",
					"error", err,
				)
				return err
			}
			recipients := make([]models.NotificationRecipient, 0, len(adminIDs))
			for _, adminID := range adminIDs {
				recipients = append(recipients, models.NotificationRecipient{
					NotificationID: notification.ID,
					AdminID:        adminID,
				})
			}
			if err := repos.Notification.CreateRecipients(ctx, recipients); err != nil {
				u.log.ErrorContext(
					ctx,
					"failed to assign notification to admins",
					"notification_id", notification.ID,
					"error", err,
				)
				return err
			}
			u.log.InfoContext(
				ctx,
				"notification created",
				"notification_id", notification.ID,
				"recipient_count", len(recipients),
			)
			return nil
		},
	)
}
func (u *NotificationUseCase) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	if id == uuid.Nil {
		return nil, domainerrors.ErrInvalidInput
	}
	notification, err := u.notificationRepo.GetByID(ctx, id)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to retrieve notification",
			"notification_id", id,
			"error", err,
		)
		return nil, err
	}
	return notification, nil
}
func (u *NotificationUseCase) ListByAdminID(ctx context.Context, adminID uint, limit int, page int, unreadOnly bool) ([]models.NotificationRecipient, int64, error) {
	if adminID == 0 {
		return nil, 0, domainerrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	recipients, total, err := u.notificationRepo.ListByAdminID(
		ctx,
		adminID,
		limit,
		page,
		unreadOnly,
	)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to list admin notifications",
			"admin_id", adminID,
			"error", err,
		)
		return nil, 0, err
	}
	return recipients, total, nil
}
func (u *NotificationUseCase) CountUnreadByAdminID(ctx context.Context, adminID uint) (int64, error) {
	if adminID == 0 {
		return 0, domainerrors.ErrInvalidInput
	}
	count, err := u.notificationRepo.CountUnreadByAdminID(ctx, adminID)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to count unread notifications",
			"admin_id", adminID,
			"error", err,
		)
		return 0, err
	}
	return count, nil
}
func (u *NotificationUseCase) MarkAsRead(ctx context.Context, adminID uint, notificationID uuid.UUID) error {
	if adminID == 0 || notificationID == uuid.Nil {
		return domainerrors.ErrInvalidInput
	}
	err := u.notificationRepo.MarkAsRead(
		ctx,
		adminID,
		notificationID,
	)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to mark notification as read",
			"admin_id", adminID,
			"notification_id", notificationID,
			"error", err,
		)
		return err
	}
	u.log.InfoContext(
		ctx,
		"notification marked as read",
		"admin_id", adminID,
		"notification_id", notificationID,
	)
	return nil
}
func (u *NotificationUseCase) MarkAllAsRead(ctx context.Context, adminID uint) error {
	if adminID == 0 {
		return domainerrors.ErrInvalidInput
	}
	err := u.notificationRepo.MarkAllAsRead(ctx, adminID)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to mark all notifications as read",
			"admin_id", adminID,
			"error", err,
		)
		return err
	}
	u.log.InfoContext(
		ctx,
		"all notifications marked as read",
		"admin_id", adminID,
	)
	return nil
}

var _ interfaces.NotificationUseCase = (*NotificationUseCase)(nil)
