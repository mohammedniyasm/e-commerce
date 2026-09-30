package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type AuthUseCase interface {
	Register(ctx context.Context, user *models.User) (*models.User, error)
	Login(ctx context.Context, email string, password string) (string, string, error)
	RefreshAccessToken(ctx context.Context,refreshToken string)(string,error)
	Logout(ctx context.Context,refreshToken string)error

	SendVerficationOTP(ctx context.Context,email string)error
	VerifyEmail(ctx context.Context,email,otp string)error
}
