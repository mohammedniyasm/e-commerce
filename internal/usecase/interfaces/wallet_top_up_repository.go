package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type WalletTopUpRepository interface {
	Create(ctx context.Context, topUp *models.WalletTopUp) error
	GetByGatewayOrderIDForUpdate(ctx context.Context, gatewayOrderID string) (*models.WalletTopUp, error)
	Update(ctx context.Context, topUp *models.WalletTopUp) error
}
