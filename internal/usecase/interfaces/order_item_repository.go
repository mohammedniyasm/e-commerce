package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type OrderItemRepository interface {
	Create(ctx context.Context, item *models.OrderItem) error
	CreateMany(ctx context.Context, items []models.OrderItem) error
	GetByID(ctx context.Context, id uint) (*models.OrderItem, error)
	ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error)
}
