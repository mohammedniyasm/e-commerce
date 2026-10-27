package handler

import (
	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notificationUseCase interfaces.NotificationUseCase
	log                 *slog.Logger
}

func NewNotificationHandler(
	notificationUseCase interfaces.NotificationUseCase,
	log *slog.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		notificationUseCase: notificationUseCase,
		log:                 log,
	}
}

func getUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch id := value.(type) {
	case int64:
		return id, id > 0

	case int:
		return int64(id), id > 0

	case uint:
		if id == 0 {
			return 0, false
		}
		return int64(id), true

	case string:
		parsedID, err := strconv.ParseInt(id, 10, 64)
		if err != nil || parsedID <= 0 {
			return 0, false
		}
		return parsedID, true

	default:
		return 0, false
	}
}
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
			Error:   "unauthorized",
		})
		return
	}
	var req request.ListNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid query parameters",
			// Error:   err.Error(),
		})
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	adminID := uint(userID)
	recipients, total, err := h.notificationUseCase.ListByAdminID(
		c.Request.Context(),
		adminID,
		req.Limit,
		req.Page,
		req.UnreadOnly,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to list notifications",
			"admin_id", adminID,
			"error", err,
		)

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "Failed to retrieve notifications",
			// Error:   "internal_server_error",
		})
		return
	}
	notifications := make([]response.NotificationResponse, 0, len(recipients))
	for _, recipient := range recipients {
		notification := recipient.Notification
		notifications = append(notifications, response.NotificationResponse{
			ID:         notification.ID,
			Type:       notification.Type,
			Title:      notification.Title,
			Message:    notification.Message,
			Priority:   notification.Priority,
			EntityType: notification.EntityType,
			EntityID:   notification.EntityID,
			EntityData: notification.EventData,
			CreatedAt:  notification.CreatedAt,
			IsRead:     recipient.IsRead,
			ReadAt:     recipient.ReadAt,
		})
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Notifications retrieved successfully",
		Data: response.NotificationListResponse{
			Notifications: notifications,
			Total:         total,
			Page:          req.Page,
			Limit:         req.Limit,
		},
	})
}
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
			Error:   "unauthorized",
		})
		return
	}
	count, err := h.notificationUseCase.CountUnreadByAdminID(
		c.Request.Context(),
		uint(userID),
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to count unread notifications",
			"admin_id", userID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "Failed to retrieve unread notification count",
			Error:   "internal_server_error",
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Unread notification count retrieved successfully",
		Data: response.UnreadNotificationCountResponse{
			Count: count,
		},
	})
}
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
			// Error:   "unauthorized",
		})
		return
	}
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil || notificationID == uuid.Nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid notification ID",
			// Error:   "invalid_notification_id",
		})
		return
	}
	err = h.notificationUseCase.MarkAsRead(
		c.Request.Context(),
		uint(userID),
		notificationID,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "Notification not found",
				// Error:   "notification_not_found",
			})
			return
		}
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to mark notification as read",
			"admin_id", userID,
			"notification_id", notificationID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "Failed to mark notification as read",
			// Error:   "internal_server_error",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Notification marked as read",
	})
}
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
			// Error:   "unauthorized",
		})
		return
	}
	err := h.notificationUseCase.MarkAllAsRead(
		c.Request.Context(),
		uint(userID),
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to mark all notifications as read",
			"admin_id", userID,
			"error", err,
		)

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "Failed to mark all notifications as read",
			// Error:   "internal_server_error",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "All notifications marked as read",
	})
}
