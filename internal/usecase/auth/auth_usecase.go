package auth

import (
	"context"
	"ecommerce/internal/delivery/http/validator"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/infrastructure/redis"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo   interfaces.UserRepository
	otpStore   redis.OTPStore
	jwtService interfaces.JWTService
	log        *slog.Logger
}

func NewAuthUseCase(userRepo interfaces.UserRepository, otpStore redis.OTPStore, jwtService interfaces.JWTService, log *slog.Logger) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		otpStore:   otpStore,
		jwtService: jwtService,
		log:        log,
	}
}

func (u *AuthUseCase) Register(ctx context.Context, user *models.User) (*models.User, error) {
	if !validator.ValidateName(user.Name) {
		return nil, domainerrors.ErrInvalidName
	}
	if !validator.ValidatePhone(user.Phone) {
		return nil, domainerrors.ErrInvalidPhone
	}
	if user.Password == nil {
		return nil, domainerrors.ErrWeakPassword
	}
	if !validator.ValidatePassword(*user.Password) {
		return nil, domainerrors.ErrWeakPassword
	}
	existingUser, err := u.userRepo.FindByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		u.log.Warn(
			"user signup failed",
			"reason", "email already exists",
			"email", user.Email,
		)
		return nil, domainerrors.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		u.log.Error(
			"failed to check existing user",
			"email", user.Email,
			"error", err,
		)
		return nil, err
	}
	if user.Role == "" {
		user.Role = models.RoleUser
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.log.Error(
			"failed to hash password",
			"error", err,
		)
		return nil, err
	}
	user.Password = stringPtr(string(hashedPassword))
	err = u.userRepo.Create(ctx, user)
	if err != nil {
		u.log.Error(
			"failed to create user",
			"email", user.Email,
			"error", err,
		)
		return nil, err
	}
	u.log.Info("user signup successful",
		"user_id", user.ID,
		"user_email", user.Email)
	return user, nil
}
func stringPtr(value string) *string {
	return &value
}
