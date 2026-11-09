package interfaces

import (
	"context"

	"github.com/google/uuid"
)

type PaymentUseCase interface {
	CreateRazorpayOrder(ctx context.Context, userID uint, orderID uuid.UUID) (string, error)
	VerifyRazorpayPayment(ctx context.Context, userID uint, orderID uuid.UUID, razorpayOrderID string, razorpayPaymentID string, signature string) error
	HandleRazorpayWebhook(ctx context.Context, payload []byte, signature string) error
}
