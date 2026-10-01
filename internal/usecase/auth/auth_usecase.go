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
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo             interfaces.UserRepository
	otpStore             redis.OTPStore
	jwtService           interfaces.JWTService
	refreshSessionStore  interfaces.RefreshSessionStore
	accessTokenBlacklist interfaces.AccessTokenBlacklist
	emailSender          interfaces.EmailSender
	log                  *slog.Logger
}

func NewAuthUseCase(userRepo interfaces.UserRepository, otpStore redis.OTPStore, jwtService interfaces.JWTService, refreshSessionStore interfaces.RefreshSessionStore, accessTokenBlacklist interfaces.AccessTokenBlacklist, emailSender interfaces.EmailSender, log *slog.Logger) *AuthUseCase {
	return &AuthUseCase{
		userRepo:             userRepo,
		otpStore:             otpStore,
		jwtService:           jwtService,
		refreshSessionStore:  refreshSessionStore,
		accessTokenBlacklist: accessTokenBlacklist,
		emailSender:          emailSender,
		log:                  log,
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
	err = u.refreshSessionStore.Save(ctx, jti, user.ID)
	if err != nil {
		u.log.Error("failed to save refreshh session",
			"uset_id", user.ID,
			"error", err,
		)
		return "", "", err
	}
	return accessToken, refreshToken, nil
}
func (u *AuthUseCase) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	userID, jti, err := u.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}
	sessionUserID, err := u.refreshSessionStore.Get(ctx, jti)
	if err != nil {
		return "", err
	}
	if userID != strconv.FormatUint(uint64(sessionUserID), 10) {
		return "", errors.New("refresh session user mismatch")
	}
	user, err := u.userRepo.FindByID(ctx, sessionUserID)
	if err != nil {
		return "", err
	}
	if user.IsBlocked {
		return "", domainerrors.ErrUserBlocked
	}
	accessToken, err := u.jwtService.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return "", err
	}
	return accessToken, nil
}
func (u *AuthUseCase) Logout(ctx context.Context, refreshToken string, accessClaims *interfaces.AccessClaims) error {
	if refreshToken == "" {
		return domainerrors.ErrInvalidCredentials
	}
	_, jti, err := u.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	errr := u.refreshSessionStore.Delete(ctx, jti)
	if errr != nil {
		return errr
	}
	remainingLifetime := time.Until(accessClaims.ExpiresAt)
	if remainingLifetime > 0 {
		err := u.accessTokenBlacklist.Blacklist(ctx, accessClaims.JTI, remainingLifetime)

		if err != nil {
			return err
		}
	}

	return nil
}
func (u *AuthUseCase) SendVerficationOTP(ctx context.Context, email string) error {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt != nil {
		return domainerrors.ErrEmailAlreadyVerified
	}
	otp, err := generateOTP()
	if err != nil {
		return err
	}
	key := emailVerificationOTPKey(email)
	err = u.otpStore.Set(ctx, key, otp, emailVerificationOTPExpiry)
	if err != nil {
		return err
	}
	err = u.emailSender.SendVerificationOTP(ctx, email, otp)
	if err != nil {
		_ = u.otpStore.Delete(ctx, key)
	}
	u.log.Info("email verification OTP generated", "email", email)
	return nil
}
func (u *AuthUseCase) ResendVerificationOTP(ctx context.Context, email string) error {
	return u.SendVerficationOTP(ctx, email)
}
func (u *AuthUseCase) VerifyEmail(ctx context.Context, userID uint, otp string) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt != nil {
		return domainerrors.ErrEmailAlreadyVerified
	}
	key := emailVerificationOTPKey(user.Email)
	storedOTP, err := u.otpStore.Get(ctx, key)
	if err != nil {
		return err
	}
	if storedOTP != otp {
		return errors.New("Invalid OTP")
	}
	verifiedAt := time.Now()
	err = u.userRepo.UpdateEmailVerifiedAt(ctx, user.ID, verifiedAt)
	if err != nil {
		return err
	}
	err = u.otpStore.Delete(ctx, key)
	if err != nil {
		return err
	}
	u.log.Info("email verified successfully", "email", user.Email)
	return nil
}
func (u *AuthUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil
		}
		return err
	}
	otp, err := generateOTP()
	if err != nil {
		return err
	}
	key := passwordResetOTPKey(user.Email)
	err = u.otpStore.Set(ctx, key, otp, passwordResetOTPExpiry)
	if err != nil {
		return err
	}
	err = u.emailSender.SendVerificationOTP(ctx, user.Email, otp)
	if err != nil {
		_ = u.otpStore.Delete(ctx, key)
		return err
	}
	u.log.Info(
		"password reset OTP sent",
		"email", email,
	)
	return nil
}
func (u *AuthUseCase) VerifyForgotPasswordOTP(ctx context.Context, email, otp string) (string, error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	key := passwordResetOTPKey(user.Email)
	storedOTP, err := u.otpStore.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if storedOTP != otp {
		return "", domainerrors.ErrInvalidOTP
	}
	resetToken, err := generaratePasswordResetToken()
	if err != nil {
		return "", err
	}
	resetKey := passwordResetTokenKey(resetToken)
	if err := u.otpStore.Set(ctx, resetKey, strconv.FormatUint(uint64(user.ID), 10), passwordResetTokenExpiry); err != nil {
		return "", err
	}
	if err := u.otpStore.Delete(ctx, key); err != nil {
		return "", err
	}
	return resetToken, nil
}
func (u *AuthUseCase) ResendForgotPasswordOTP(ctx context.Context, email string) error {
	return u.ForgotPassword(ctx, email)
}
func (u *AuthUseCase) ResetPassword(ctx context.Context, resetToken string, newPassword string) error {
	if resetToken == "" || newPassword == "" {
		return domainerrors.ErrInvalidCredentials
	}
	if !validator.ValidatePassword(newPassword) {
		return domainerrors.ErrWeakPassword
	}
	key := passwordResetTokenKey(resetToken)
	userIDString, err := u.otpStore.Get(ctx, key)
	if err != nil {
		return err
	}
	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		return err
	}
	user, err := u.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := u.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return err
	}
	err = u.otpStore.Delete(ctx, key)
	if err != nil {
		return err
	}
	u.log.Info(
		"password reset successful",
		"user_id", user.ID,
	)
	return nil
}
func stringPtr(value string) *string {
	return &value
}
