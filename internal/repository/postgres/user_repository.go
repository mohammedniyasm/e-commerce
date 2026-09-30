package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}
func (r *UserRepository) FindByID(ctx context.Context, ID uint) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, ID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}
func (r *UserRepository) UpdateEmailVerifiedAt(ctx context.Context, userID uint, verifiedAt time.Time) error {
	result := r.db.WithContext(ctx).Table("users").Where("id = ?", userID).Update("email_verified_at", verifiedAt)
	return result.Error
}
func (r *UserRepository) IsBlocked(ctx context.Context, userID uint) (bool, error) {
	var isBlocked bool
	result := r.db.WithContext(ctx).Table("users").Select("is_blocked").Where("id = ?", userID).Scan(&isBlocked)
	if result.Error != nil {
		return false, result.Error
	}
	return isBlocked, nil
}
