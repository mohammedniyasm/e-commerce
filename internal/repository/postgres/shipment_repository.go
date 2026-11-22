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
func (r *ShipmentRepository) List(ctx context.Context, filter models.ShipmentListFilter) ([]models.Shipment, int64, error) {
	var shipments []models.Shipment
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Shipment{})

	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		query = query.
			Joins("LEFT JOIN orders ON orders.id = shipments.order_id").
			Where(
				"(CAST(shipments.id AS TEXT) ILIKE ? OR shipments.order_id::text ILIKE ? OR COALESCE(shipments.courier_name, '') ILIKE ? OR COALESCE(shipments.tracking_number, '') ILIKE ? OR COALESCE(orders.order_number, '') ILIKE ?)",
				pattern,
				pattern,
				pattern,
				pattern,
				pattern,
			)
	}
	if filter.ShippingStatus != nil {
		query = query.Where("shipments.shipping_status = ?", *filter.ShippingStatus)
	}
	if filter.Type != nil {
		query = query.Where("shipments.type = ?", *filter.Type)
	}
	if filter.CourierName != "" {
		query = query.Where("COALESCE(shipments.courier_name, '') ILIKE ?", "%"+filter.CourierName+"%")
	}
	if filter.OrderID != nil {
		query = query.Where("shipments.order_id = ?", *filter.OrderID)
	}
	if filter.ReturnRequestID != nil {
		query = query.Where("shipments.return_request_id = ?", *filter.ReturnRequestID)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("shipments.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("shipments.created_at < ?", filter.CreatedTo.AddDate(0, 0, 1))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortOrders := map[string]string{
		"newest":       "shipments.created_at DESC, shipments.id DESC",
		"oldest":       "shipments.created_at ASC, shipments.id ASC",
		"updated_desc": "shipments.updated_at DESC, shipments.id DESC",
		"updated_asc":  "shipments.updated_at ASC, shipments.id ASC",
		"status_asc":   "shipments.shipping_status ASC, shipments.created_at DESC",
		"status_desc":  "shipments.shipping_status DESC, shipments.created_at DESC",
	}
	orderBy, ok := sortOrders[filter.Sort]
	if !ok {
		orderBy = sortOrders["newest"]
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	err := query.
		Preload("Order", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "order_number")
		}).
		Order(orderBy).
		Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Find(&shipments).Error
	if err != nil {
		return nil, 0, err
	}
	return shipments, total, nil
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
	if err := r.db.WithContext(ctx).Where("order_id = ? AND type = ?", orderID, models.ShipmentTypeDelivery).First(&shipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &shipment, nil
}
func (r *ShipmentRepository) Update(ctx context.Context, shipment *models.Shipment) error {
	result := r.db.WithContext(ctx).Model(&models.Shipment{}).Where("id = ?", shipment.ID).Select(
		"CourierName",
		"TrackingNumber",
		"UpdatedAt",
	).Updates(shipment)
	if result.Error != nil {
		return result.Error
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
func (r *ShipmentRepository) GetByReturnRequestID(ctx context.Context, returnRequestID uint) ([]models.Shipment, error) {
	var shipments []models.Shipment
	err := r.db.WithContext(ctx).
		Preload("History").
		Where("return_request_id = ?", returnRequestID).
		Order("created_at ASC").
		Find(&shipments).Error
	if err != nil {
		return nil, err
	}
	return shipments, nil
}
func (r *ShipmentRepository) GetByReturnRequestIDAndType(ctx context.Context, returnRequestID uint, shipmentType models.ShipmentType) (*models.Shipment, error) {
	var shipment models.Shipment
	err := r.db.WithContext(ctx).
		Where(
			"return_request_id = ? AND type = ?",
			returnRequestID,
			shipmentType,
		).
		First(&shipment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &shipment, nil
}

var _ interfaces.ShipmentRepository = (*ShipmentRepository)(nil)
