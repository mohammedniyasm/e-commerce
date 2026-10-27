package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id uint) (*models.Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.Payment, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*models.Payment, error)
	Update(ctx context.Context, payment *models.Payment) error
}
