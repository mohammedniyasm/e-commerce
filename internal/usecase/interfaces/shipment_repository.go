package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type ShipmentRepository interface {
	Create(ctx context.Context, shipment *models.Shipment) error
	GetByID(ctx context.Context, id uint) (*models.Shipment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Shipment, error)
	GetByReturnRequestID(ctx context.Context, returnRequestID uint) ([]models.Shipment, error)
	Update(ctx context.Context, shipment *models.Shipment) error
	UpdateShippingStatus(ctx context.Context, shipmentID uint, status string, description *string, location *string) error
	GetTrackingHistory(ctx context.Context, shipmentID uint) ([]models.ShipmentTrackingHistory, error)
	GetByReturnRequestIDAndType(ctx context.Context, returnRequestID uint, shipmentType models.ShipmentType) (*models.Shipment, error)
}
