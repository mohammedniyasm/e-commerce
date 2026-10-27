package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error)

	//user
	ListByUserID(ctx context.Context, userID uint, limit int, page int, sort *string) ([]models.Order, int64, error)
	//admin
	List(ctx context.Context, userID *uint, search string, limit int, page int, sort *string, orderStatus *string, paymentStatus *string) ([]models.Order, int64, error)

	Update(ctx context.Context, order *models.Order) error
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, note *string) error
	ListExpiredPendingOrders(ctx context.Context, now time.Time, limit int) ([]models.Order, error)
	LockByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
}
