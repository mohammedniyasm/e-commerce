package postgres

import (
	"context"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletTopUpRepository struct {
	db *gorm.DB
}

func NewWalletTopUpRepository(db *gorm.DB) *walletTopUpRepository {
	return &walletTopUpRepository{
		db: db,
	}
}

func (r *walletTopUpRepository) Create(ctx context.Context, topUp *models.WalletTopUp) error {
	return r.db.WithContext(ctx).Create(topUp).Error
}
func (r *walletTopUpRepository) GetByGatewayOrderIDForUpdate(ctx context.Context, gatewayOrderID string) (*models.WalletTopUp, error) {
	var topUp models.WalletTopUp
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("gateway_order_id = ?", gatewayOrderID).
		First(&topUp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &topUp, nil
}
func (r *walletTopUpRepository) Update(ctx context.Context, topUp *models.WalletTopUp) error {
	result := r.db.WithContext(ctx).
		Model(&models.WalletTopUp{}).
		Where("id = ?", topUp.ID).
		Select(
			"GatewayPaymentID",
			"Status",
			"PaidAt",
			"UpdatedAt",
		).
		Updates(topUp)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}

var _ interfaces.WalletTopUpRepository = (*walletTopUpRepository)(nil)
