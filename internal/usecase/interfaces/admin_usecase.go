package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type AdminUseCase interface {
	ListUsers(ctx context.Context, search string, page int, limit int) ([]models.User, int64, error)
	BlockUser(ctx context.Context, userID uint) error
	UnblockUser(ctx context.Context, userID uint) error
	AddCustomer(ctx context.Context, name string, email string, phone string, password string) (*models.User, error)
	GetCustomer(ctx context.Context, userID uint) (*models.User, error)
	UpdateCustomer(ctx context.Context, userID uint, name string, email string, phone string) (*models.User, error)
	DeleteCustomer(ctx context.Context, userID uint) error
}
