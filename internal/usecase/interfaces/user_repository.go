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
	IsBlocked(ctx context.Context, userID uint) (bool, error)
	FindByGoogleID(ctx context.Context, googleID string) (*models.User, error)

	ListUsers(ctx context.Context, search string, offset int, limit int) ([]models.User, int64, error)

	UpdateEmailVerifiedAt(ctx context.Context, userID uint, verifiedAt time.Time) error
	UpdateProfileImage(ctx context.Context, userID uint, profileImage string) error
	UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error
	UpdateProfile(ctx context.Context, userID uint, name, phone string) error
	LinkGoogleID(ctx context.Context, userID uint, googleID string) error
	UpdateEmail(ctx context.Context, userID uint, email string) error
	SetBlocked(ctx context.Context, userID uint, blocked bool) error

	Delete(ctx context.Context, userID uint) error
}
