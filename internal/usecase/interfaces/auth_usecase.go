package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type AuthUseCase interface {
	Register(ctx context.Context, user *models.User) (*models.User, error)
	Login(ctx context.Context, email string, password string) (string, string, error)
	RefreshAccessToken(ctx context.Context,refreshToken string)(string,error)
	Logout(ctx context.Context,refreshToken string,accessClaims *AccessClaims)error

	SendVerficationOTP(ctx context.Context,email string)error
	ResendVerificationOTP(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context,userID uint,otp string)error

	ForgotPassword(ctx context.Context, email string) error
	VerifyForgotPasswordOTP(ctx context.Context, email,otp string) (string,error)
	ResendForgotPasswordOTP(ctx context.Context, email string) error
	ResetPassword(ctx context.Context,resetToken string,newPassword string)error
}
