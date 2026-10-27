package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type OrderUseCase interface {
	GetByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error)
	ListByUserID(ctx context.Context, userID uint, limit int, page int, sort *string) ([]models.Order, int64, error)
	// Admin order management
	List(ctx context.Context, userID *uint, search string, limit int, page int, sort *string, orderStatus *string, paymentStatus *string) ([]models.Order, int64, error)
	// Order status management
	PlaceOrder(ctx context.Context, userID uint, cartItemIDs []uint, addressID uint, paymentMethod string, couponCode *string) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, note *string) error
	CancelOrder(ctx context.Context, userID uint, orderID uuid.UUID) error
	UpdateOrderAddress(ctx context.Context, userID uint, orderID uuid.UUID, address models.OrderAddress) error
	GenerateInvoice(ctx context.Context, userID uint, orderID uuid.UUID) ([]byte, error)
	UpdateCODPaymentStatus(ctx context.Context, orderID uuid.UUID, status string) error
}
