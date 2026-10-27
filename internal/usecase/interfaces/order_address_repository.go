package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type OrderAddressRepository interface {
	Create(ctx context.Context, address *models.OrderAddress) error
	GetByID(ctx context.Context, id uint) (*models.OrderAddress, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.OrderAddress, error)
	Update(ctx context.Context, address *models.OrderAddress) error
}
