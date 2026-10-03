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
	googleTokenVerifier  interfaces.GoogleTokenVerifier
	log                  *slog.Logger
}

func NewAuthUseCase(userRepo interfaces.UserRepository, otpStore redis.OTPStore, jwtService interfaces.JWTService, refreshSessionStore interfaces.RefreshSessionStore, accessTokenBlacklist interfaces.AccessTokenBlacklist, emailSender interfaces.EmailSender, googleTokenVerifier interfaces.GoogleTokenVerifier, log *slog.Logger) *AuthUseCase {
	return &AuthUseCase{
		userRepo:             userRepo,
		otpStore:             otpStore,
		jwtService:           jwtService,
		refreshSessionStore:  refreshSessionStore,
		accessTokenBlacklist: accessTokenBlacklist,
		googleTokenVerifier:  googleTokenVerifier,
		emailSender:          emailSender,

		log: log,
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
func (u *AuthUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	return u.login(ctx, email, password, false)
}
func (u *AuthUseCase) AdminLogin(ctx context.Context, email, password string) (string, string, error) {
	return u.login(ctx, email, password, true)
}
func (u *AuthUseCase) login(ctx context.Context, email string, password string, adminOnly bool) (string, string, error) {
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
		u.log.Error(
			"failed to find user during login",
			"error", err,
		)
		return "", "", err
	}
	if user.Password == nil {
		return "", "", domainerrors.ErrInvalidCredentials
	}
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password))
	if err != nil {
		u.log.Warn("user login failed",
			"email", email,
			"error", "invalid credentials",
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
	if adminOnly && user.Role != models.RoleAdmin {
		u.log.Warn(
			"admin login failed",
			"user_id", user.ID,
			"reason", "user is not admin",
		)
		return "", "", domainerrors.ErrInvalidCredentials
	}
	accessToken, err := u.jwtService.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		u.log.Error("failed to generate access token",
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
			"user_id", user.ID,
			"error", err,
		)
		return "", "", err
	}
	u.log.Info(
		"user login successful",
		"user_id", user.ID,
		"admin_login", adminOnly,
	)
	return accessToken, refreshToken, nil
}
func (u *AuthUseCase) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	userID, jti, err := u.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		u.log.Warn(
			"refresh token validation failed",
			"error", err,
		)
		return "", "", err
	}
	sessionUserID, found, err := u.refreshSessionStore.Consume(ctx, jti)
	if err != nil {
		u.log.Error(
			"failed to consume refresh session",
			"error", err,
		)
		return "", "", err
	}
	if !found {
		return "", "", domainerrors.ErrInvalidCredentials
	}
	if userID != strconv.FormatUint(uint64(sessionUserID), 10) {
		u.log.Warn(
			"refresh session user mismatch",
			"session_user_id", sessionUserID,
		)
		return "", "", errors.New("refresh session user mismatch")
	}
	user, err := u.userRepo.FindByID(ctx, sessionUserID)
	if err != nil {
		u.log.Error(
			"failed to find user during token refresh",
			"user_id", sessionUserID,
			"error", err,
		)
		return "", "", err
	}
	if user.IsBlocked {
		u.log.Warn(
			"refresh denied for blocked user",
			"user_id", user.ID,
		)
		return "", "", domainerrors.ErrUserBlocked
	}
	accessToken, err := u.jwtService.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		u.log.Error(
			"failed to generate access token",
			"user_id", user.ID,
			"error", err,
		)
		return "", "", err
	}
	newRefreshToken, newJTI, err := u.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		u.log.Error(
			"failed to generate refresh token",
			"user_id", user.ID,
			"error", err,
		)
		return "", "", err
	}
	err = u.refreshSessionStore.Save(ctx, newJTI, user.ID)
	if err != nil {
		u.log.Error(
			"failed to save rotated refresh session",
			"user_id", user.ID,
			"error", err,
		)
		return "", "", err
	}
	u.log.Info(
		"refresh token rotated",
		"user_id", user.ID,
	)
	return accessToken, newRefreshToken, nil
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
		u.log.Error(
			"failed to delete refresh session during logout",
			"error", errr,
		)
		return errr
	}
	remainingLifetime := time.Until(accessClaims.ExpiresAt)
	if remainingLifetime > 0 {
		err := u.accessTokenBlacklist.Blacklist(ctx, accessClaims.JTI, remainingLifetime)
		if err != nil {
			u.log.Error(
				"failed to blacklist access token during logout",
				"user_id", accessClaims.UserID,
				"error", err,
			)
			return err
		}
	}
	u.log.Info(
		"user logout successful",
		"user_id", accessClaims.UserID,
	)
	return nil
}
func (u *AuthUseCase) SendVerficationOTP(ctx context.Context, email string) error {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		u.log.Error(
			"failed to find user for email verification OTP",
			"error", err,
		)
		return err
	}
	if user.EmailVerifiedAt != nil {
		return domainerrors.ErrEmailAlreadyVerified
	}
	otp, err := generateOTP()
	if err != nil {
		u.log.Error(
			"failed to generate email verification OTP",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	key := emailVerificationOTPKey(email)
	err = u.otpStore.Set(ctx, key, otp, emailVerificationOTPExpiry)
	if err != nil {
		u.log.Error(
			"failed to store email verification OTP",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	err = u.emailSender.SendVerificationOTP(ctx, email, otp)
	if err != nil {
		u.log.Error(
			"failed to send email verification OTP",
			"user_id", user.ID,
			"error", err,
		)
		_ = u.otpStore.Delete(ctx, key)
	}
	u.log.Info(
		"email verification OTP sent successfully",
		"user_id", user.ID,
	)
	return nil
}
func (u *AuthUseCase) ResendVerificationOTP(ctx context.Context, email string) error {
	return u.SendVerficationOTP(ctx, email)
}
func (u *AuthUseCase) VerifyEmail(ctx context.Context, userID uint, otp string) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		u.log.Error(
			"failed to find user for email verification",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	if user.EmailVerifiedAt != nil {
		return domainerrors.ErrEmailAlreadyVerified
	}
	key := emailVerificationOTPKey(user.Email)
	storedOTP, err := u.otpStore.Get(ctx, key)
	if err != nil {
		u.log.Error(
			"failed to retrieve email verification OTP",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	if storedOTP != otp {
		u.log.Warn(
			"email verification failed",
			"user_id", user.ID,
			"reason", "invalid OTP",
		)
		return errors.New("Invalid OTP")
	}
	verifiedAt := time.Now()
	err = u.userRepo.UpdateEmailVerifiedAt(ctx, user.ID, verifiedAt)
	if err != nil {
		u.log.Error(
			"failed to update email verification status",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	err = u.otpStore.Delete(ctx, key)
	if err != nil {
		u.log.Error(
			"failed to delete email verification OTP",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"email verified successfully",
		"user_id", user.ID,
	)
	return nil
}
func (u *AuthUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil
		}
		u.log.Error(
			"failed to find user for password reset",
			"error", err,
		)
		return err
	}
	otp, err := generateOTP()
	if err != nil {
		u.log.Error(
			"failed to generate password reset OTP",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	key := passwordResetOTPKey(user.Email)
	err = u.otpStore.Set(ctx, key, otp, passwordResetOTPExpiry)
	if err != nil {
		u.log.Error(
			"failed to store password reset OTP",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	err = u.emailSender.SendVerificationOTP(ctx, user.Email, otp)
	if err != nil {
		u.log.Error(
			"failed to send password reset OTP",
			"user_id", user.ID,
			"error", err,
		)
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
		u.log.Error(
			"failed to find user for password reset OTP verification",
			"error", err,
		)
		return "", err
	}
	key := passwordResetOTPKey(user.Email)
	storedOTP, err := u.otpStore.Get(ctx, key)
	if err != nil {
		u.log.Error(
			"failed to retrieve password reset OTP",
			"user_id", user.ID,
			"error", err,
		)

		return "", err
	}
	if storedOTP != otp {
		u.log.Warn(
			"password reset OTP verification failed",
			"user_id", user.ID,
			"reason", "invalid OTP",
		)
		return "", domainerrors.ErrInvalidOTP
	}
	resetToken, err := generaratePasswordResetToken()
	if err != nil {
		u.log.Error(
			"failed to generate password reset token",
			"user_id", user.ID,
			"error", err,
		)
		return "", err
	}
	resetKey := passwordResetTokenKey(resetToken)
	if err := u.otpStore.Set(ctx, resetKey, strconv.FormatUint(uint64(user.ID), 10), passwordResetTokenExpiry); err != nil {
		u.log.Error(
			"failed to store password reset token",
			"user_id", user.ID,
			"error", err,
		)
		return "", err
	}
	if err := u.otpStore.Delete(ctx, key); err != nil {
		u.log.Error(
			"failed to delete password reset OTP",
			"user_id", user.ID,
			"error", err,
		)
		return "", err
	}
	u.log.Info(
		"password reset OTP verified successfully",
		"user_id", user.ID,
	)
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
		u.log.Error(
			"failed to retrieve password reset token",
			"error", err,
		)
		return err
	}
	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		u.log.Error(
			"failed to parse user ID from password reset token",
			"error", err,
		)
		return err
	}
	user, err := u.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		u.log.Error(
			"failed to find user during password reset",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		u.log.Error(
			"failed to hash password during password reset",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	if err := u.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		u.log.Error(
			"failed to update password during password reset",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	err = u.otpStore.Delete(ctx, key)
	if err != nil {
		u.log.Error(
			"failed to delete password reset token",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"password reset successful",
		"user_id", user.ID,
	)
	return nil
}
func (u *AuthUseCase) GoogleLogin(ctx context.Context, idToken string) (string, string, error) {
	googleUser, err := u.googleTokenVerifier.Verify(ctx, idToken)
	if err != nil {
		u.log.Warn(
			"google login failed",
			"reason", "invalid google token",
			"error", err,
		)
		return "", "", domainerrors.ErrInvalidCredentials
	}
	user, err := u.userRepo.FindByGoogleID(ctx, googleUser.SubjectID)
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		u.log.Error(
			"failed to find user by Google ID",
			"error", err,
		)
		return "", "", err
	}
	if err == nil && user != nil {
		if user.IsBlocked {
			u.log.Warn(
				"google login failed",
				"user_id", user.ID,
				"reason", "user is blocked",
			)
			return "", "", domainerrors.ErrUserBlocked
		}
		return u.generateAuthTokens(ctx, user)
	}
	user, err = u.userRepo.FindByEmail(ctx, googleUser.Email)
	if err == nil && user != nil {
		if err := u.userRepo.LinkGoogleID(ctx, user.ID, googleUser.SubjectID); err != nil {
			u.log.Error(
				"failed to link Google account",
				"user_id", user.ID,
				"error", err,
			)
			return "", "", err
		}
		if user.EmailVerifiedAt == nil {
			now := time.Now()
			if err := u.userRepo.UpdateEmailVerifiedAt(ctx, user.ID, now); err != nil {
				u.log.Error(
					"failed to verify email for Google account",
					"user_id", user.ID,
					"error", err,
				)
				return "", "", err
			}
		}
		if user.IsBlocked {
			u.log.Warn(
				"google login failed",
				"user_id", user.ID,
				"reason", "user is blocked",
			)
			return "", "", domainerrors.ErrUserBlocked
		}
		u.log.Info(
			"Google account linked to existing user",
			"user_id", user.ID,
		)
		return u.generateAuthTokens(ctx, user)
	}
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		u.log.Error(
			"failed to find user by email during Google login",
			"error", err,
		)
		return "", "", err
	}
	googleID := googleUser.SubjectID
	now := time.Now()

	newUser := &models.User{
		Name:            googleUser.Name,
		Email:           googleUser.Email,
		Password:        nil,
		GoogleID:        &googleID,
		EmailVerifiedAt: &now,
		ProfileImage:    googleUser.Picture,
		Role:            models.RoleUser,
		IsBlocked:       false,
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		u.log.Error(
			"failed to create Google user",
			"error", err,
		)
		return "", "", err
	}

	u.log.Info(
		"google user created",
		"user_id", newUser.ID,
		"email", newUser.Email,
	)

	return u.generateAuthTokens(ctx, newUser)
}
func stringPtr(value string) *string {
	return &value
}
func (u *AuthUseCase) generateAuthTokens(ctx context.Context, user *models.User) (string, string, error) {
	accessToken, err := u.jwtService.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		u.log.Error(
			"failed to generate access token",
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
	if err := u.refreshSessionStore.Save(ctx, jti, user.ID); err != nil {
		u.log.Error(
			"failed to save refresh session",
			"user_id", user.ID,
			"error", err,
		)
		return "", "", err
	}
	u.log.Info(
		"authentication tokens generated successfully",
		"user_id", user.ID,
	)
	return accessToken, refreshToken, nil
}
