package postgres

import (
	"context"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderStatusHistoryRepository struct {
	db *gorm.DB
}

func NewOrderStatusHistoryRepository(
	db *gorm.DB,
) *OrderStatusHistoryRepository {
	return &OrderStatusHistoryRepository{
		db: db,
	}
}

func (r *OrderStatusHistoryRepository) Create(ctx context.Context, history *models.OrderStatusHistory) error {
	return r.db.
		WithContext(ctx).
		Create(history).
		Error
}
func (r *OrderStatusHistoryRepository) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.OrderStatusHistory, error) {
	var history []models.OrderStatusHistory
	err := r.db.
		WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC, id ASC").
		Find(&history).
		Error

	if err != nil {
		return nil, err
	}

	return history, nil
}

var _ interfaces.OrderStatusHistoryRepository = (*OrderStatusHistoryRepository)(nil)
