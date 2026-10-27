package response

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationResponse struct {
	ID         uuid.UUID        `json:"id"`
	Type       string           `json:"type"`
	Title      string           `json:"title"`
	Message    string           `json:"message"`
	Priority   string           `json:"priority"`
	EntityType *string          `json:"entity_type,omitempty"`
	EntityID   *string          `json:"entity_id,omitempty"`
	EntityData *json.RawMessage `json:"entity_data,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	IsRead     bool             `json:"is_read"`
	ReadAt     *time.Time       `json:"read_at,omitempty"`
}
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	Limit         int                    `json:"limit"`
}
type UnreadNotificationCountResponse struct {
	Count int64 `json:"count"`
}
type NotificationMessageResponse struct {
	Message string `json:"message"`
}
