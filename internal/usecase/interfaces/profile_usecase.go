package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
	"io"
)

type ProfileUseCase interface {
	GetProfile(ctx context.Context, userID uint) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uint, name string, phone string) (*models.User, error)
	ChangePassword(ctx context.Context, userID uint, currentPassword string, newPassword string) error
	SendEmailChangeOTP(ctx context.Context, userID uint, newEmail string) error
	VerifyEmailChangeOTP(ctx context.Context, userID uint, otp string) error
	UploadProfileImage(ctx context.Context, userID uint, file io.Reader, size int64, contentType string) (string, error)
	DeleteProfileImage(ctx context.Context,userID uint) error
}
