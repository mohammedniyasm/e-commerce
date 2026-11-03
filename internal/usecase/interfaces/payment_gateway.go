package interfaces

type PaymentGateway interface {
	CreateOrder(amount int64, receipt string) (string, error)
	VerifyPaymentSignature(orderID string, paymentID string, signature string) error
	FetchPayment(paymentID string) (map[string]interface{}, error)
}
