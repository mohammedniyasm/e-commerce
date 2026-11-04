package request

import "github.com/google/uuid"

type PaymentVerifyRequest struct {
	OrderID           uuid.UUID `json:"order_id" binding:"required"`
	RazorpayOrderID   string    `json:"razorpay_order_id" binding:"required"`
	RazorpayPaymentID string    `json:"razorpay_payment_id" binding:"required"`
	RazorpaySignature string    `json:"razorpay_signature" binding:"required"`
}

type CreatePaymentOrderRequest struct {
	OrderID uuid.UUID `json:"order_id" binding:"required"`
}
