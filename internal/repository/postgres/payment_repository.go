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

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}
func (r *PaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}
func (r *PaymentRepository) GetByID(ctx context.Context, id uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &payment, nil
}
func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}
func (r *PaymentRepository) GetByTransactionID(ctx context.Context, transactionID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Where("transaction_id = ?", transactionID).
		First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &payment, nil
}
func (r *PaymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	result := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ?", payment.ID).
		Select(
			"PaymentMethod",
			"PaymentStatus",
			"GatewayOrderID",
			"TransactionID",
			"GatewayResponse",
			"FailureReason",
			"Amount",
			"PaidAt",
			"UpdatedAt",
		).
		Updates(payment)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}

var _ interfaces.PaymentRepository = (*PaymentRepository)(nil)
