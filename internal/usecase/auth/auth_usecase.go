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
	userRepo            interfaces.UserRepository
	otpStore            redis.OTPStore
	jwtService          interfaces.JWTService
	refreshSessionStore interfaces.RefreshSessionStore
	log                 *slog.Logger
}

func NewAuthUseCase(userRepo interfaces.UserRepository, otpStore redis.OTPStore, jwtService interfaces.JWTService, refreshSessionStore interfaces.RefreshSessionStore, log *slog.Logger) *AuthUseCase {
	return &AuthUseCase{
		userRepo:            userRepo,
		otpStore:            otpStore,
		jwtService:          jwtService,
		refreshSessionStore: refreshSessionStore,
		log:                 log,
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
func (u *AuthUseCase) Login(ctx context.Context, email string, password string) (string, string, error) {
	if email == "" || password == "" {
		return "", "", domainerrors.ErrInvalidCredentials
	}
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			u.log.Warn("user login failed",
				"email", email,
				"error", domainerrors.ErrUserNotFound,
			)
			return "", "", domainerrors.ErrInvalidCredentials
		}
		return "", "", err
	}
	if user.Password == nil {
		return "", "", domainerrors.ErrInvalidCredentials
	}
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password))
	if err != nil {
		u.log.Warn("user login failed",
			"email", email,
			"error", "invalid password",
		)
		return "", "", domainerrors.ErrInvalidCredentials
	}
	if user.IsBlocked {
		u.log.Warn("user login failed",
			"email", email,
			"reason", "user is blocked",
		)
		return "", "", domainerrors.ErrUserBlocked
	}
	accessToken, err := u.jwtService.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		u.log.Warn("failed to generate access token",
			"user_id", user.ID,
			"error", err,
		)
		return "", "", err
	}
	refreshToken, jti, err := u.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		u.log.Error(
			"failed to generate refresh token",
			"user_id", user.ID,
			"error", err,
		)
		return "", "", err
	}
	err=u.refreshSessionStore.Save(ctx,jti,user.ID)
	if err != nil{
		u.log.Error("failed to save refreshh session",
			"uset_id",user.ID,
			"error",err,
	)
	return "","",err
	}
	return accessToken, refreshToken, nil
}
func stringPtr(value string) *string {
	return &value
}
