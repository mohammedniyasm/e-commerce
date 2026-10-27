package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID         uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type       string           `gorm:"not null;size:50;index"`
	Title      string           `gorm:"not null;size:255"`
	Message    string           `gorm:"type:text;not null"`
	Priority   string           `gorm:"not null;size:20;index"`
	EntityType *string          `gorm:"size:50"`
	EntityID   *string          `gorm:"size:100"`
	EventData  *json.RawMessage `gorm:"type:jsonb"`
	CreatedAt  time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
}
