package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type ReturnRequestRepository interface {
	Create(ctx context.Context, request *models.ReturnRequest) error
	GetByID(ctx context.Context, id uint) (*models.ReturnRequest, error)
	GetByIDAndUserID(ctx context.Context, id uint, userID uint) (*models.ReturnRequest, error)
	ListByUserID(ctx context.Context, userID uint, limit int, page int) ([]models.ReturnRequest, int64, error)
	List(ctx context.Context, orderID *uuid.UUID, status *string, limit int, page int) ([]models.ReturnRequest, int64, error)
	Update(ctx context.Context, request *models.ReturnRequest) error
	HasActiveRequest(ctx context.Context, orderItemID uint) (bool, error)
}
