package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	domainmodels "ecommerce/internal/domain/models"
	postgresmodels "ecommerce/internal/repository/postgres/models"
	"errors"

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
func (r *UserRepository) Create(ctx context.Context, user *domainmodels.User) error {
	dbUser := &postgresmodels.User{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Password:  user.Password,
		Role:      string(user.Role),
		IsBlocked: user.IsBlocked,
		LastSeen:  user.LastSeen,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	result := r.db.WithContext(ctx).Create(dbUser)
	if result.Error != nil {
		return result.Error
	}
	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt
	return nil
}
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domainmodels.User, error) {
	var dbUser postgresmodels.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&dbUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, result.Error
	}
	user := &domainmodels.User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Email:     dbUser.Email,
		Phone:     dbUser.Phone,
		Password:  dbUser.Password,
		Role:      domainmodels.UserRole(dbUser.Role),
		IsBlocked: dbUser.IsBlocked,
		LastSeen:  dbUser.LastSeen,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
	return user, nil
}
