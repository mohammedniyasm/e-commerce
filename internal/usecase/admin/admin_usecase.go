package admin

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"ecommerce/internal/delivery/http/validator"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"golang.org/x/crypto/bcrypt"
)

type AdminUseCase struct {
	userRepo interfaces.UserRepository
	log      *slog.Logger
}

func NewAdminUseCase(
	userRepo interfaces.UserRepository,
	log *slog.Logger,
) *AdminUseCase {
	return &AdminUseCase{
		userRepo: userRepo,
		log:      log,
	}
}

func (u *AdminUseCase) ListUsers(ctx context.Context, search string, page int, limit int) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	users, total, err := u.userRepo.ListUsers(ctx, search, offset, limit)
	if err != nil {
		u.log.Error(
			"failed to list users",
			"search", search,
			"page", page,
			"limit", limit,
			"error", err,
		)

		return nil, 0, err
	}
	return users, total, nil
}
func (u *AdminUseCase) BlockUser(ctx context.Context, userID uint) error {
	if userID == 0 {
		return domainerrors.ErrUserNotFound
	}
	if err := u.userRepo.SetBlocked(ctx, userID, true); err != nil {
		u.log.Error(
			"failed to block user",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"user blocked",
		"user_id", userID,
	)
	return nil
}
func (u *AdminUseCase) UnblockUser(ctx context.Context, userID uint) error {
	if userID == 0 {
		return domainerrors.ErrUserNotFound
	}
	if err := u.userRepo.SetBlocked(ctx, userID, false); err != nil {
		u.log.Error(
			"failed to unblock user",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"user unblocked",
		"user_id", userID,
	)
	return nil
}
func (u *AdminUseCase) AddCustomer(ctx context.Context, name string, email string, phone string, password string) (*models.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	phone = strings.TrimSpace(phone)
	if name == "" || email == "" || phone == "" || password == "" {
		return nil, domainerrors.ErrInvalidCredentials
	}
	if !validator.ValidateName(name) {
		return nil, domainerrors.ErrInvalidName
	}
	if !validator.ValidatePhone(phone) {
		return nil, domainerrors.ErrInvalidPhone
	}
	if !validator.ValidatePassword(password) {
		return nil, domainerrors.ErrWeakPassword
	}
	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, domainerrors.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}
	hashed := string(hashedPassword)
	user := &models.User{
		Name:         name,
		Email:        email,
		Phone:        phone,
		Password:     &hashed,
		Role:         models.RoleUser,
		IsBlocked:    false,
		ProfileImage: "",
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		u.log.Error(
			"failed to create customer",
			"email", email,
			"error", err,
		)
		return nil, err
	}
	u.log.Info(
		"customer created by admin",
		"user_id", user.ID,
		"email", user.Email,
	)
	return user, nil
}
func (u *AdminUseCase) GetCustomer(ctx context.Context, userID uint) (*models.User, error) {

	if userID == 0 {
		return nil, domainerrors.ErrUserNotFound
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Role != models.RoleUser {
		return nil, domainerrors.ErrUserNotFound
	}

	return user, nil
}
func (u *AdminUseCase) UpdateCustomer(ctx context.Context, userID uint, name string, email string, phone string) (*models.User, error) {

	if userID == 0 {
		return nil, domainerrors.ErrUserNotFound
	}

	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	phone = strings.TrimSpace(phone)

	if name == "" || email == "" || phone == "" {
		return nil, domainerrors.ErrInvalidCredentials
	}
	if !validator.ValidateName(name) {
		return nil, domainerrors.ErrInvalidName
	}
	if !validator.ValidatePhone(phone) {
		return nil, domainerrors.ErrInvalidPhone
	}
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Role != models.RoleUser {
		return nil, domainerrors.ErrUserNotFound
	}

	if email != user.Email {
		existingUser, err := u.userRepo.FindByEmail(ctx, email)

		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, domainerrors.ErrEmailAlreadyExists
		}

		if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, err
		}
	}

	if err := u.userRepo.UpdateProfile(
		ctx,
		userID,
		name,
		phone,
	); err != nil {
		u.log.Error(
			"failed to update customer profile",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}

	if email != user.Email {
		if err := u.userRepo.UpdateEmail(
			ctx,
			userID,
			email,
		); err != nil {
			u.log.Error(
				"failed to update customer email",
				"user_id", userID,
				"error", err,
			)
			return nil, err
		}
	}

	updatedUser, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	u.log.Info(
		"customer updated by admin",
		"user_id", userID,
	)

	return updatedUser, nil
}
func (u *AdminUseCase) DeleteCustomer(ctx context.Context,userID uint) error {

	if userID == 0 {
		return domainerrors.ErrUserNotFound
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Role != models.RoleUser {
		return domainerrors.ErrUserNotFound
	}

	if err := u.userRepo.Delete(ctx, userID); err != nil {
		u.log.Error(
			"failed to delete customer",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	u.log.Info(
		"customer deleted by admin",
		"user_id", userID,
	)

	return nil
}
