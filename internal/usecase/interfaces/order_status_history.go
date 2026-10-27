package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type OrderStatusHistoryRepository interface {
	Create(ctx context.Context, history *models.OrderStatusHistory) error
	ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.OrderStatusHistory, error)
}
