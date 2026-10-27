package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type ShipmentUseCase interface {
	CreateShipment(ctx context.Context, shipment *models.Shipment) error
	GetShipmentByID(ctx context.Context, id uint) (*models.Shipment, error)
	GetShipmentByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Shipment, error)
	UpdateShipment(ctx context.Context, shipment *models.Shipment) error
	UpdateShippingStatus(ctx context.Context, shipmentID uint, status string, description *string, location *string) error
	GetTrackingHistory(ctx context.Context, shipmentID uint) ([]models.ShipmentTrackingHistory, error)
	GetCustomerShipment(ctx context.Context, userID uint, orderID uuid.UUID) (*models.Shipment, error)
	GetCustomerTrackingHistory(ctx context.Context, userID uint, orderID uuid.UUID) ([]models.ShipmentTrackingHistory, error)
}
