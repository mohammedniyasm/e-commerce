package auth

import (
	"context"
	"ecommerce/internal/delivery/http/validator"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo interfaces.UserRepository
}

func NewAuthUseCase(userRepo interfaces.UserRepository) *AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
	}
}

func (u *AuthUseCase) Register(ctx context.Context, user *models.User) (*models.User, error) {
	if !validator.ValidateName(user.Name) {
		return nil, domainerrors.ErrInvalidName
	}
	if !validator.ValidatePhone(user.Phone) {
		return nil, domainerrors.ErrInvalidPhone
	}
	if !validator.ValidatePassword(*user.Password) {
		return nil, domainerrors.ErrWeakPassword
	}
	existingUser, err := u.userRepo.FindByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return nil, domainerrors.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}
	if user.Role == "" {
		user.Role = models.RoleUser
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = stringPtr(string(hashedPassword))
	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func stringPtr(value string) *string {
	return &value
}
