package postgres

import (
	"context"
	"errors"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShipmentRepository struct {
	db *gorm.DB
}

func NewShipmentRepository(db *gorm.DB) *ShipmentRepository {
	return &ShipmentRepository{
		db: db,
	}
}
func (r *ShipmentRepository) Create(ctx context.Context, shipment *models.Shipment) error {
	return r.db.WithContext(ctx).Create(shipment).Error
}
func (r *ShipmentRepository) GetByID(ctx context.Context, id uint) (*models.Shipment, error) {
	var shipment models.Shipment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&shipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &shipment, nil
}
func (r *ShipmentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Shipment, error) {
	var shipment models.Shipment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&shipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &shipment, nil
}
func (r *ShipmentRepository) Update(ctx context.Context, shipment *models.Shipment) error {
	result := r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", shipment.ID).Select(
		"CourierName",
		"TrackingNumber",
		"UpdatedAt",
	).Updates(shipment)
	if result.Error != nil {
		return nil
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *ShipmentRepository) UpdateShippingStatus(ctx context.Context, shipmentID uint, status string, description *string, location *string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var shipment models.Shipment
		if err := tx.Where("id = ?", shipmentID).First(&shipment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerrors.ErrRecordNotFound
			}
			return err
		}
		if err := tx.Model(&shipment).Update("shipping_status", status).Error; err != nil {
			return err
		}
		history := models.ShipmentTrackingHistory{
			ShipmentID:     shipmentID,
			ShippingStatus: status,
			Description:    description,
			Location:       location,
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		switch status {
		case "in_transit":
			if err := tx.Model(&shipment).Update("shipped_at", time.Now()).Error; err != nil {
				return err
			}
		case "delivered":
			if err := tx.Model(&shipment).Update("delivered_at", time.Now()).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *ShipmentRepository) GetTrackingHistory(ctx context.Context, shipmentID uint) ([]models.ShipmentTrackingHistory, error) {
	var history []models.ShipmentTrackingHistory
	if err := r.db.WithContext(ctx).Where("shipment_id = ?", shipmentID).Order("created_at DESC").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

var _ interfaces.ShipmentRepository = (*ShipmentRepository)(nil)
