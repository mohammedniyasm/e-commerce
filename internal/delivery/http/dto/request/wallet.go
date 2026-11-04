package request

type AdminWalletCreditRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Reason string  `json:"reason" binding:"required"`
}
type CreateTopUpOrderRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type VerifyTopUpPaymentRequest struct {
	GatewayOrderID   string `json:"gateway_order_id" binding:"required"`
	GatewayPaymentID string `json:"gateway_payment_id" binding:"required"`
	Signature        string `json:"razorpay_signature" binding:"required"`
}
