package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}
func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(order).Error
}
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).Where("id = ?", id).
		Preload("Items").
		Preload("Address").
		Preload("Payments").
		Preload("StatusHistory").
		Preload("Shipment").
		Preload("Shipment.History").
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &order, nil
}
func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).Where("order_number = ?", orderNumber).
		Preload("Items").
		Preload("Address").
		Preload("Payments").
		Preload("StatusHistory").
		Preload("Shipment").
		Preload("Shipment.History").
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &order, nil
}
func (r *OrderRepository) ListByUserID(ctx context.Context, userID uint, limit int, page int, sort *string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if sort != nil {
		switch *sort {
		case "new":
			query = query.Order("created_at DESC")
		case "old":
			query = query.Order("created_at ASC")
		default:
			query = query.Order("created_at DESC")
		}
	} else {
		query = query.Order("created_at DESC")
	}

	offset := (page - 1) * limit

	err := query.
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
func (r *OrderRepository) List(ctx context.Context, userID *uint, search string, limit int, page int, sort *string, orderStatus *string, paymentStatus *string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Order{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("order_number ILIKE ?", searchPattern)
	}
	if orderStatus != nil {
		switch *orderStatus {
		case "pending",
			"processing",
			"shipped",
			"delivered",
			"cancelled",
			"returned",
			"confirmed":
			query = query.Where("order_status = ?", *orderStatus)
		}
	}
	if paymentStatus != nil {
		switch *paymentStatus {
		case "pending",
			"paid",
			"failed",
			"refunded",
			"partially_refunded",
			"confirmed":
			query = query.Where("payment_status = ?", *paymentStatus)
		}
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if sort != nil {
		switch *sort {
		case "new":
			query = query.Order("created_at DESC")
		case "old":
			query = query.Order("created_at ASC")
		default:
			query = query.Order("created_at DESC")
		}
	} else {
		query = query.Order("created_at DESC")
	}
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
func (r *OrderRepository) Update(ctx context.Context, order *models.Order) error {
	err := r.db.WithContext(ctx).Where("id = ?", order.ID).
		Select(
			"Subtotal",
			"DiscountAmount",
			"ShippingAmount",
			"OrderStatus",
			"PaymentStatus",
			"TotalAmount",
			"UpdatedAt",
		).Updates(order)
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, note *string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerrors.ErrRecordNotFound
			}
			return err
		}
		if err := tx.Model(&order).Update("order_status", status).Error; err != nil {
			return err
		}
		history := models.OrderStatusHistory{
			OrderID: orderID,
			Status:  status,
			Note:    note,
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		return nil
	})
}
func (r *OrderRepository) ListExpiredPendingOrders(ctx context.Context, now time.Time, limit int) ([]models.Order, error) {
	var orders []models.Order
	if limit <= 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	err := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("orders.order_status = ?", "pending").
		Where("orders.payment_status IN ?", []string{"pending", "failed"}).
		Where("orders.payment_deadline IS NOT NULL").
		Where("orders.payment_deadline <= ?", now).
		Where(`
			EXISTS (
				SELECT 1
				FROM payments p
				WHERE p.order_id = orders.id
				  AND p.id = (
					  SELECT MAX(p2.id)
					  FROM payments p2
					  WHERE p2.order_id = orders.id
				  )
				  AND LOWER(p.payment_method) NOT IN (?, ?)
			)
		`, "cod", "cash_on_delivery").
		Order("orders.payment_deadline ASC").
		Limit(limit).
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}
func (r *OrderRepository) LockByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	var order models.Order

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", orderID).
		First(&order).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}

		return nil, err
	}

	return &order, nil
}

var _ interfaces.OrderRepository = (*OrderRepository)(nil)
