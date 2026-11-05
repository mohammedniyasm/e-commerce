package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"
	"encoding/hex"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	orderRepo      interfaces.OrderRepository
	paymentRepo    interfaces.PaymentRepository
	paymentGateway interfaces.PaymentGateway
	unitOfWork     interfaces.UnitOfWork
	razorpayKeyID  string
	webHookSecret  string
}

func NewPaymentUseCase(
	orderRepo interfaces.OrderRepository,
	paymentRepo interfaces.PaymentRepository,
	paymentGateway interfaces.PaymentGateway,
	unitOfWork interfaces.UnitOfWork,
	razorpayKeyID string,
	webHookSecret string,
) *PaymentUseCase {
	return &PaymentUseCase{
		orderRepo:      orderRepo,
		paymentRepo:    paymentRepo,
		paymentGateway: paymentGateway,
		unitOfWork:     unitOfWork,
		razorpayKeyID:  razorpayKeyID,
		webHookSecret:  webHookSecret,
	}
}
func (u *PaymentUseCase) CreateRazorpayOrder(ctx context.Context, userID uint, orderID uuid.UUID) (string, error) {
	if userID <= 0 {
		return "", domainerrors.ErrInvalidUserID
	}
	if orderID == uuid.Nil {
		return "", domainerrors.ErrInvalidOrderID
	}
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return "", err
	}
	if order.UserID != userID {
		return "", domainerrors.ErrRecordNotFound
	}
	if strings.ToLower(order.OrderStatus) != "pending" ||
		strings.ToLower(order.PaymentStatus) != "pending" {
		return "", domainerrors.ErrInvalidInput
	}
	deadline := order.PlacedAt.Add(10 * time.Minute)
	if time.Now().After(deadline) {
		return "", domainerrors.ErrInvalidInput
	}
	payments, err := u.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return "", err
	}
	if len(payments) == 0 {
		return "", domainerrors.ErrRecordNotFound
	}
	payment := &payments[0]
	if strings.ToLower(payment.PaymentStatus) != "pending" {
		return "", domainerrors.ErrInvalidInput
	}
	if payment.GatewayOrderID != nil && *payment.GatewayOrderID != "" {
		return *payment.GatewayOrderID, nil
	}
	amount := int64(math.Round(payment.Amount * 100))
	if amount <= 0 {
		return "", domainerrors.ErrInvalidInput
	}
	gatewayOrderID, err := u.paymentGateway.CreateOrder(amount, order.OrderNumber)
	if err != nil {
		return "", err
	}
	payment.GatewayOrderID = &gatewayOrderID
	payment.UpdatedAt = time.Now()
	if err := u.paymentRepo.Update(ctx, payment); err != nil {
		return "", err
	}
	return gatewayOrderID, nil
}
func (u *PaymentUseCase) VerifyRazorpayPayment(ctx context.Context, userID uint, orderID uuid.UUID, razorpayOrderID string, razorpayPaymentID string, signature string) error {
	if userID <= 0 {
		return domainerrors.ErrInvalidUserID
	}
	if orderID == uuid.Nil {
		return domainerrors.ErrInvalidOrderID
	}
	if razorpayOrderID == "" ||
		razorpayPaymentID == "" ||
		signature == "" {
		return domainerrors.ErrInvalidInput
	}
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.UserID != userID {
		return domainerrors.ErrRecordNotFound
	}
	if order.OrderStatus != "pending" ||
		order.PaymentStatus != "pending" {
		return domainerrors.ErrInvalidInput
	}
	deadline := order.PlacedAt.Add(10 * time.Minute)
	if time.Now().After(deadline) {
		return domainerrors.ErrInvalidInput
	}
	payments, err := u.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	if len(payments) == 0 {
		return domainerrors.ErrRecordNotFound
	}
	payment := &payments[0]
	if payment.PaymentStatus != "pending" {
		return domainerrors.ErrInvalidInput
	}
	if err := u.paymentGateway.VerifyPaymentSignature(razorpayOrderID, razorpayPaymentID, signature); err != nil {
		return err
	}
	gatewayPayment, err := u.paymentGateway.FetchPayment(razorpayPaymentID)
	if err != nil {
		return err
	}
	gatewayOrderID, ok := gatewayPayment["order_id"].(string)
	if !ok || gatewayOrderID != razorpayOrderID {
		return domainerrors.ErrInvalidInput
	}
	currency, ok := gatewayPayment["currency"].(string)
	if !ok || currency != "INR" {
		return domainerrors.ErrInvalidInput
	}
	status, ok := gatewayPayment["status"].(string)
	if !ok || status != "captured" {
		return domainerrors.ErrInvalidInput
	}
	amount, ok := gatewayPayment["amount"].(float64)
	if !ok || int64(amount) != int64(math.Round(payment.Amount*100)) {
		return domainerrors.ErrInvalidInput
	}
	gatewayResponse, err := json.Marshal(gatewayPayment)
	if err != nil {
		return err
	}
	now := time.Now()
	return u.unitOfWork.WithinTransaction(ctx, func(repos interfaces.TransactionRepositories) error {
		lockedOrder, err := repos.Order.LockByID(ctx, orderID)
		if err != nil {
			return err
		}
		if lockedOrder.UserID != userID {
			return domainerrors.ErrRecordNotFound
		}
		if lockedOrder.PaymentStatus == "paid" {
			return nil
		}
		if lockedOrder.PaymentStatus != "pending" {
			return domainerrors.ErrInvalidInput
		}
		txPayments, err := repos.Payment.GetByOrderID(ctx, orderID)
		if err != nil {
			return err
		}
		if len(txPayments) == 0 {
			return domainerrors.ErrRecordNotFound
		}
		txPayment := txPayments[0]
		if txPayment.PaymentStatus != "pending" ||
			txPayment.GatewayOrderID == nil ||
			*txPayment.GatewayOrderID != razorpayOrderID {
			return domainerrors.ErrInvalidInput
		}
		transactionID := razorpayPaymentID
		txPayment.PaymentStatus = "paid"
		txPayment.TransactionID = &transactionID
		txPayment.GatewayResponse = gatewayResponse
		txPayment.PaidAt = &now
		txPayment.UpdatedAt = now
		if err := repos.Payment.Update(ctx, &txPayment); err != nil {
			return err
		}
		lockedOrder.PaymentStatus = "paid"
		lockedOrder.OrderStatus = "confirmed"
		lockedOrder.UpdatedAt = now
		return repos.Order.Update(ctx, lockedOrder)
	})
}
func (u *PaymentUseCase) HandleRazorpayWebhook(ctx context.Context, payload []byte, signature string) error {
	if len(payload) == 0 || signature == "" {
		return domainerrors.ErrInvalidInput
	}
	mac := hmac.New(sha256.New, []byte(u.webHookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal(
		[]byte(expectedSignature),
		[]byte(signature),
	) {
		return domainerrors.ErrInvalidInput
	}
	var event struct {
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					ID               string `json:"id"`
					OrderID          string `json:"order_id"`
					Amount           int64  `json:"amount"`
					Currency         string `json:"currency"`
					Status           string `json:"status"`
					ErrorDescription string `json:"error_description"`
				} `json:"entity"`
			} `json:"payment"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}
	if event.Event != "payment.captured" &&
		event.Event != "payment.failed" {
		return nil
	}
	gatewayPayment := event.Payload.Payment.Entity
	if gatewayPayment.ID == "" || gatewayPayment.OrderID == "" {
		return domainerrors.ErrInvalidInput
	}
	if gatewayPayment.Currency != "INR" {
		return domainerrors.ErrInvalidInput
	}
	switch event.Event {
	case "payment.captured":
		if gatewayPayment.Status != "captured" {
			return domainerrors.ErrInvalidInput
		}
	case "payment.failed":
		if gatewayPayment.Status != "failed" {
			return domainerrors.ErrInvalidInput
		}
	}
	now := time.Now()
	return u.unitOfWork.WithinTransaction(
		ctx,
		func(repos interfaces.TransactionRepositories) error {
			txPayment, err := repos.Payment.GetByGatewayOrderID(
				ctx,
				gatewayPayment.OrderID,
			)
			if err != nil {
				return err
			}
			if txPayment == nil {
				return domainerrors.ErrRecordNotFound
			}
			lockedOrder, err := repos.Order.LockByID(
				ctx,
				txPayment.OrderID,
			)
			if err != nil {
				return err
			}
			txPayment, err = repos.Payment.GetByGatewayOrderIDForUpdate(
				ctx,
				gatewayPayment.OrderID,
			)
			if err != nil {
				return err
			}
			if txPayment.OrderID != lockedOrder.ID {
				return domainerrors.ErrInvalidInput
			}
			if txPayment.GatewayOrderID == nil ||
				*txPayment.GatewayOrderID != gatewayPayment.OrderID {
				return domainerrors.ErrInvalidInput
			}
			expectedAmount := int64(math.Round(txPayment.Amount * 100))
			if gatewayPayment.Amount != expectedAmount {
				return domainerrors.ErrInvalidInput
			}
			if txPayment.PaymentStatus == "paid" ||
				txPayment.PaymentStatus == "refunded" {
				return nil
			}
			if lockedOrder.PaymentStatus == "refunded" ||
				lockedOrder.PaymentStatus == "partially_refunded" {
				return nil
			}
			if event.Event == "payment.failed" &&
				txPayment.PaymentStatus == "failed" {
				return nil
			}
			response, err := json.Marshal(gatewayPayment)
			if err != nil {
				return err
			}
			transactionID := gatewayPayment.ID
			txPayment.TransactionID = &transactionID
			txPayment.GatewayResponse = response
			txPayment.UpdatedAt = now
			switch event.Event {
			case "payment.captured":
				txPayment.PaymentStatus = "paid"
				txPayment.PaidAt = &now
				txPayment.FailureReason = nil
				lockedOrder.PaymentStatus = "paid"
			case "payment.failed":
				if txPayment.PaymentStatus == "paid" {
					return nil
				}
				txPayment.PaymentStatus = "failed"
				if gatewayPayment.ErrorDescription != "" {
					reason := gatewayPayment.ErrorDescription
					txPayment.FailureReason = &reason
				}
				lockedOrder.PaymentStatus = "failed"
			}
			lockedOrder.UpdatedAt = now
			lockedOrder.OrderStatus = "confirmed"
			if err := repos.Payment.Update(ctx, txPayment); err != nil {
				return err
			}
			return repos.Order.Update(ctx, lockedOrder)
		},
	)
}
