package interfaces

import "context"

type EmailSender interface {
	SendVerificationOTP(ctx context.Context,to string,otp string)error
}