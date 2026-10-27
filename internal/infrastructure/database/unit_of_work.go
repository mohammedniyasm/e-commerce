package database

import (
	"context"
	"ecommerce/internal/repository/postgres"
	"ecommerce/internal/usecase/interfaces"

	"gorm.io/gorm"
)

type GormUnitOfWork struct {
	db *gorm.DB
}

func NewGormUnitOfWork(db *gorm.DB) *GormUnitOfWork {
	return &GormUnitOfWork{
		db: db,
	}
}
func (uow *GormUnitOfWork) WithinTransaction(ctx context.Context, fn func(repos interfaces.TransactionRepositories) error) error {
	return uow.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repos := interfaces.TransactionRepositories{
			Order:              postgres.NewOrderRepository(tx),
			OrderItem:          postgres.NewOrderItemRepository(tx),
			OrderAddress:       postgres.NewOrderAddressRepository(tx),
			OrderStatusHistory: postgres.NewOrderStatusHistoryRepository(tx),
			Shipment:           postgres.NewShipmentRepository(tx),
			Address:            postgres.NewAddressRepository(tx),
			ProductVariant:     postgres.NewProductVariantRepository(tx),
			Coupon:             postgres.NewCouponRepository(tx),
			CouponUsage:        postgres.NewCouponUsageRepository(tx),
			Cart:               postgres.NewCartRepository(tx),
			Payment:            postgres.NewPaymentRepository(tx),
			Notification:       postgres.NewNotificationRepository(tx),
			User:               postgres.NewUserRepository(tx),
			ReturnRequest:      postgres.NewReturnRequestRepository(tx),
		}
		return fn(repos)
	})
}

var _ interfaces.UnitOfWork = (*GormUnitOfWork)(nil)
