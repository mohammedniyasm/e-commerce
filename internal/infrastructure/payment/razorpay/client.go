package razorpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/razorpay/razorpay-go"
)

type Client struct {
	client    *razorpay.Client
	keySecret string
}

func NewClient(keyID, keySecret string) (*Client, error) {
	if keyID == "" || keySecret == "" {
		return nil, fmt.Errorf("razorpay creadentials are required")
	}
	client := razorpay.NewClient(keyID, keySecret)
	return &Client{
		client:    client,
		keySecret: keySecret,
	}, nil
}
func (c *Client) CreateOrder(amount int64, receipt string) (string, error) {
	if amount <= 0 {
		return "", errors.New("payment amount must be greater than zero")
	}

	if receipt == "" || len(receipt) > 40 {
		return "", errors.New("invalid receipt")
	}
	data := map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  receipt,
	}
	result, err := c.client.Order.Create(data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create razorpay order: %w", err)
	}
	orderID, ok := result["id"].(string)
	if !ok || orderID == "" {
		return "", errors.New("invalid razorpay order response")
	}
	return orderID, nil
}
func (c *Client) VerifyPaymentSignature(orderID string, paymentID string, signature string) error {
	if orderID == "" || paymentID == "" || signature == "" {
		return errors.New("invalid payment verification details")
	}
	payload := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(c.keySecret))
	mac.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal(
		[]byte(expectedSignature),
		[]byte(signature),
	) {
		return errors.New("invalid payment signature")
	}
	return nil
}
func (c *Client) FetchPayment(paymentID string) (map[string]interface{}, error) {
	if paymentID == "" {
		return nil, errors.New("payment ID is required")
	}
	result, err := c.client.Payment.Fetch(paymentID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch razorpay payment: %w", err)
	}
	return result, nil
}
func VerifyWebhookSignature(body []byte, signature string, webhookSecret string) bool {
	if len(body) == 0 || signature == "" || webhookSecret == "" {
		return false
	}
	receivedSignature, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	expectedSignature := mac.Sum(nil)
	return subtle.ConstantTimeCompare(
		expectedSignature,
		receivedSignature,
	) == 1
}
