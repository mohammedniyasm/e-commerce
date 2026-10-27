package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type ReturnRequestUseCase interface {
	Create(ctx context.Context, userID uint, orderID uuid.UUID, orderItemID uint, requestType string, requestedVariantID *uint, quantity int, reason string) (*models.ReturnRequest, error)
	GetByID(ctx context.Context, requestID uint, userID uint) (*models.ReturnRequest, error)
	ListByUserID(ctx context.Context, userID uint, limit int, page int) ([]models.ReturnRequest, int64, error)
	List(ctx context.Context, orderID *uuid.UUID, status *string, limit int, page int) ([]models.ReturnRequest, int64, error)
	UpdateStatus(ctx context.Context, returnRequestID uint, status string, adminNote *string) (*models.ReturnRequest, error)
}
