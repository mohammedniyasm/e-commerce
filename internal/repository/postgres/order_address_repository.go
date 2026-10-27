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

type OrderAddressRepository struct {
	db *gorm.DB
}

func NewOrderAddressRepository(db *gorm.DB) *OrderAddressRepository {
	return &OrderAddressRepository{
		db: db,
	}
}
func (r *OrderAddressRepository) Create(ctx context.Context, address *models.OrderAddress) error {
	return r.db.WithContext(ctx).Create(address).Error
}
func (r *OrderAddressRepository) GetByID(ctx context.Context, id uint) (*models.OrderAddress, error) {
	var address models.OrderAddress

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&address).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}

	return &address, nil
}
func (r *OrderAddressRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.OrderAddress, error) {
	var address models.OrderAddress

	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		First(&address).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}

	return &address, nil
}
func (r *OrderAddressRepository) Update(ctx context.Context, address *models.OrderAddress) error {
	result := r.db.WithContext(ctx).
		Model(&models.OrderAddress{}).
		Where("id = ?", address.ID).
		Select("*").
		Updates(address)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}

	return nil
}

var _ interfaces.OrderAddressRepository = (*OrderAddressRepository)(nil)
