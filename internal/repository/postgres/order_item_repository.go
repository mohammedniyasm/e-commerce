package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) *OrderItemRepository {
	return &OrderItemRepository{
		db: db,
	}
}
func (r *OrderItemRepository) Create(ctx context.Context, item *models.OrderItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *OrderItemRepository) CreateMany(ctx context.Context, items []models.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}
func (r *OrderItemRepository) GetByID(ctx context.Context, id uint) (*models.OrderItem, error) {
	var item models.OrderItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}
func (r *OrderItemRepository) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error) {
	var items []models.OrderItem
	if err := r.db.WithContext(ctx).Order("id ASC").Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

var _ interfaces.OrderItemRepository = (*OrderItemRepository)(nil)
