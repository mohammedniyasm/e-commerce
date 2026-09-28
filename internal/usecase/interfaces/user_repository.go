package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context,id uint)(*models.User,error)
}
