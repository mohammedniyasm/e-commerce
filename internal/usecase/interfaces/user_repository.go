package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	UpdateEmailVerifiedAt(ctx context.Context, userID uint, verifiedAt time.Time) error
	IsBlocked(ctx context.Context,userID uint)(bool,error)
}
