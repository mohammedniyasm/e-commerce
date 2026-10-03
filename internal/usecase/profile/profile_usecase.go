package profile

import (
	"context"
	"ecommerce/internal/delivery/http/validator"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ProfileUseCase struct {
	userRepo    interfaces.UserRepository
	otpStore    interfaces.OTPstore
	emailSender interfaces.EmailSender
	storage     interfaces.ObjectStorage
	log         *slog.Logger
}

func NewProfileUseCase(
	userRepo interfaces.UserRepository,
	otpStore interfaces.OTPstore,
	emailSender interfaces.EmailSender,
	storage interfaces.ObjectStorage,
	log *slog.Logger,
) *ProfileUseCase {
	return &ProfileUseCase{
		userRepo:    userRepo,
		otpStore:    otpStore,
		emailSender: emailSender,
		storage:     storage,
		log:         log,
	}
}

func (u *ProfileUseCase) GetProfile(ctx context.Context, userID uint) (*models.User, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		u.log.Error(
			"failed to get user profile",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}
	return user, err
}
func (u *ProfileUseCase) UpdateProfile(ctx context.Context, userID uint, name string, phone string) (*models.User, error) {

	if !validator.ValidateName(name) {
		return nil, domainerrors.ErrInvalidName
	}

	if !validator.ValidatePhone(phone) {
		return nil, domainerrors.ErrInvalidPhone
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		u.log.Error(
			"failed to get user for profile update",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}

	err = u.userRepo.UpdateProfile(
		ctx,
		userID,
		name,
		phone,
	)
	if err != nil {
		u.log.Error(
			"failed to update user profile",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}

	user.Name = name
	user.Phone = phone

	u.log.Info(
		"user profile updated",
		"user_id", userID,
	)

	return user, nil
}
func (u *ProfileUseCase) ChangePassword(ctx context.Context, userID uint, currentPassword string, newPassword string) error {

	if userID == 0 {
		return domainerrors.ErrInvalidUserID
	}

	if currentPassword == "" || newPassword == "" {
		return domainerrors.ErrInvalidCredentials
	}

	if !validator.ValidatePassword(newPassword) {
		return domainerrors.ErrWeakPassword
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		u.log.Error(
			"failed to get user for password change",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	if user.Password == nil {
		return domainerrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(*user.Password),
		[]byte(currentPassword),
	); err != nil {
		return domainerrors.ErrInvalidCredentials
	}

	if currentPassword == newPassword {
		return domainerrors.ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		u.log.Error(
			"failed to hash new password",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	if err := u.userRepo.UpdatePassword(
		ctx,
		userID,
		string(hashedPassword),
	); err != nil {
		u.log.Error(
			"failed to update password",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	u.log.Info(
		"password changed successfully",
		"user_id", userID,
	)

	return nil
}
func (u *ProfileUseCase) SendEmailChangeOTP(ctx context.Context, userID uint, newEmail string) error {
	if userID == 0 {
		return domainerrors.ErrInvalidUserID
	}
	newEmail = strings.TrimSpace(strings.ToLower(newEmail))
	if newEmail == "" {
		return domainerrors.ErrInvalidCredentials
	}
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		u.log.Error(
			"failed to get user for email change",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	if strings.EqualFold(user.Email, newEmail) {
		return domainerrors.ErrEmailAlreadyExists
	}
	existingUser, err := u.userRepo.FindByEmail(ctx, newEmail)
	if err == nil && existingUser != nil {
		return domainerrors.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		u.log.Error(
			"failed to check email availability",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	otp, err := generateOTP()
	if err != nil {
		u.log.Error(
			"failed to generate email change OTP",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	key := emailChangeOTPKey(userID)

	if err := u.otpStore.Set(
		ctx,
		key,
		otp,
		emailChangeOTPExpiry,
	); err != nil {
		u.log.Error(
			"failed to store email change OTP",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	emailKey := emailChangePendingEmailKey(userID)

	if err := u.otpStore.Set(
		ctx,
		emailKey,
		newEmail,
		emailChangeOTPExpiry,
	); err != nil {
		u.log.Error(
			"failed to store pending email",
			"user_id", userID,
			"error", err,
		)
		_ = u.otpStore.Delete(ctx, key)
		return err
	}
	u.log.Info(
		"email change OTP sending",
		"user_id", userID,
	)
	if err := u.emailSender.SendVerificationOTP(
		ctx,
		newEmail,
		otp,
	); err != nil {
		u.log.Error(
			"failed to send email change OTP",
			"user_id", userID,
			"error", err,
		)
		_ = u.otpStore.Delete(ctx, key)
		_ = u.otpStore.Delete(ctx, emailKey)
		return err
	}
	u.log.Info(
		"email change OTP sent",
		"user_id", userID,
	)
	return nil
}
func (u *ProfileUseCase) VerifyEmailChangeOTP(ctx context.Context, userID uint, otp string) error {
	if userID == 0 || strings.TrimSpace(otp) == "" {
		return domainerrors.ErrInvalidCredentials
	}

	otpKey := emailChangeOTPKey(userID)
	emailKey := emailChangePendingEmailKey(userID)

	storedOTP, err := u.otpStore.Get(ctx, otpKey)
	if err != nil {
		u.log.Error(
			"failed to retrieve email change OTP",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	if storedOTP != otp {
		return domainerrors.ErrInvalidOTP
	}

	pendingEmail, err := u.otpStore.Get(ctx, emailKey)
	if err != nil {
		u.log.Error(
			"failed to retrieve pending email",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	pendingEmail = strings.TrimSpace(strings.ToLower(pendingEmail))

	if pendingEmail == "" {
		return domainerrors.ErrInvalidCredentials
	}
	existingUser, err := u.userRepo.FindByEmail(ctx, pendingEmail)
	if err == nil && existingUser != nil {
		if existingUser.ID != userID {
			return domainerrors.ErrEmailAlreadyExists
		}
		return domainerrors.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		u.log.Error(
			"failed to check pending email availability",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	now := time.Now()
	if err := u.userRepo.UpdateEmail(
		ctx,
		userID,
		pendingEmail,
	); err != nil {
		u.log.Error(
			"failed to update user email",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	if err := u.userRepo.UpdateEmailVerifiedAt(
		ctx,
		userID,
		now,
	); err != nil {
		u.log.Error(
			"failed to update user emailverifiedAt",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	if err := u.otpStore.Delete(ctx, otpKey); err != nil {
		u.log.Error(
			"failed to delete email change OTP",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	if err := u.otpStore.Delete(ctx, emailKey); err != nil {
		u.log.Error(
			"failed to delete pending email",
			"user_id", userID,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"email changed successfully",
		"user_id", userID,
	)
	return nil
}
func (u *ProfileUseCase) UploadProfileImage(ctx context.Context, userID uint, file io.Reader, size int64, contentType string) (string, error) {
	if userID == 0 {
		return "", domainerrors.ErrInvalidUserID
	}
	const maxFileSize int64 = 5 * 1024 * 1024
	if size <= 0 || size > maxFileSize {
		return "", domainerrors.ErrInvalidProfileImage
	}
	switch contentType {
	case "image/jpeg",
		"image/png",
		"image/webp":
	default:
		return "", domainerrors.ErrInvalidProfileImage
	}
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	// ext := ".jpg"
	// switch contentType {
	// case "image/png":
	// 	ext = ".png"
	// case "image/webp":
	// 	ext = ".webp"
	// }
	objectKey := fmt.Sprintf(
		"profiles/%d/profile",
		userID,
	)
	imageURL, err := u.storage.Upload(
		ctx,
		objectKey,
		file,
		size,
		contentType,
	)
	if err != nil {
		u.log.Error(
			"failed to upload profile image",
			"user_id", userID,
			"error", err,
		)
		return "", err
	}

	if err := u.userRepo.UpdateProfileImage(
		ctx,
		userID,
		imageURL,
	); err != nil {
		_ = u.storage.Delete(ctx, objectKey)

		u.log.Error(
			"failed to save profile image",
			"user_id", userID,
			"error", err,
		)

		return "", err
	}
	u.log.Info(
		"profile image updated",
		"user_id", userID,
	)
	_ = user
	return imageURL, nil
}
func (u *ProfileUseCase) DeleteProfileImage(ctx context.Context, userID uint) error {

	if userID == 0 {
		return domainerrors.ErrInvalidUserID
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.ProfileImage == "" {
		return nil
	}

	objectKey := fmt.Sprintf("profiles/%d/profile", userID)

	if err := u.storage.Delete(ctx, objectKey); err != nil {
		u.log.ErrorContext(
			ctx,
			"profile image delete failed",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	if err := u.userRepo.UpdateProfileImage(
		ctx,
		userID,
		"",
	); err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to clear profile image",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	u.log.InfoContext(
		ctx,
		"profile image deleted",
		"user_id", userID,
	)

	return nil
}
