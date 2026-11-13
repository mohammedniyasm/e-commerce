package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	domainErrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testRazorpayKeyID   = "rzp_test_key"
	testWebhookSecret   = "whsec_test_secret"
	testGatewayOrderID  = "order_rzp_123"
	testGatewayPayID    = "pay_rzp_456"
	testRazorpaySig     = "sig_valid"
)

type paymentFixture struct {
	uc        *PaymentUseCase
	orderRepo *mocks.MockOrderRepository
	payRepo   *mocks.MockPaymentRepository
	gateway   *mocks.MockPaymentGateway
	uow       *mocks.MockUnitOfWork
	txOrder   *mocks.MockOrderRepository
	txPay     *mocks.MockPaymentRepository
}

func newPaymentFixture(t *testing.T) *paymentFixture {
	t.Helper()
	orderRepo := mocks.NewMockOrderRepository(t)
	payRepo := mocks.NewMockPaymentRepository(t)
	gateway := mocks.NewMockPaymentGateway(t)
	uow := mocks.NewMockUnitOfWork(t)
	txOrder := mocks.NewMockOrderRepository(t)
	txPay := mocks.NewMockPaymentRepository(t)

	return &paymentFixture{
		uc:        NewPaymentUseCase(orderRepo, payRepo, gateway, uow, testRazorpayKeyID, testWebhookSecret),
		orderRepo: orderRepo,
		payRepo:   payRepo,
		gateway:   gateway,
		uow:       uow,
		txOrder:   txOrder,
		txPay:     txPay,
	}
}

func (f *paymentFixture) expectTX(ctx context.Context) {
	f.uow.EXPECT().WithinTransaction(ctx, mock.AnythingOfType("func(interfaces.TransactionRepositories) error")).
		RunAndReturn(func(_ context.Context, fn func(interfaces.TransactionRepositories) error) error {
			return fn(interfaces.TransactionRepositories{
				Order:   f.txOrder,
				Payment: f.txPay,
			})
		}).Once()
}

func validPendingOrder(userID uint, orderID uuid.UUID) *models.Order {
	return &models.Order{
		ID:            orderID,
		UserID:        userID,
		OrderNumber:   "ORD-1001",
		OrderStatus:   "pending",
		PaymentStatus: "pending",
		PlacedAt:      time.Now().Add(-2 * time.Minute),
	}
}

func signWebhookPayload(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(testWebhookSecret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestPaymentUseCase_CreateRazorpayOrder(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	userID := uint(11)

	t.Run("invalid identifiers", func(t *testing.T) {
		f := newPaymentFixture(t)
		got, err := f.uc.CreateRazorpayOrder(ctx, 0, orderID)
		assert.Empty(t, got)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidUserID)

		got, err = f.uc.CreateRazorpayOrder(ctx, userID, uuid.Nil)
		assert.Empty(t, got)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidOrderID)
	})

	t.Run("order not found or owned by another user", func(t *testing.T) {
		f := newPaymentFixture(t)
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(nil, domainErrors.ErrRecordNotFound).Once()
		got, err := f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		assert.Empty(t, got)
		assert.ErrorIs(t, err, domainErrors.ErrRecordNotFound)

		otherOrder := validPendingOrder(99, orderID)
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(otherOrder, nil).Once()
		got, err = f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		assert.Empty(t, got)
		assert.ErrorIs(t, err, domainErrors.ErrRecordNotFound)
	})

	t.Run("invalid order status or expired deadline", func(t *testing.T) {
		tests := []struct {
			name  string
			order *models.Order
		}{
			{
				name: "order not pending",
				order: &models.Order{
					ID: orderID, UserID: userID, OrderStatus: "confirmed", PaymentStatus: "pending", PlacedAt: time.Now(),
				},
			},
			{
				name: "payment not pending",
				order: &models.Order{
					ID: orderID, UserID: userID, OrderStatus: "pending", PaymentStatus: "paid", PlacedAt: time.Now(),
				},
			},
			{
				name: "payment window expired",
				order: &models.Order{
					ID: orderID, UserID: userID, OrderStatus: "pending", PaymentStatus: "pending", PlacedAt: time.Now().Add(-15 * time.Minute),
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				f := newPaymentFixture(t)
				f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(tc.order, nil).Once()

				got, err := f.uc.CreateRazorpayOrder(ctx, userID, orderID)
				assert.Empty(t, got)
				assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
			})
		}
	})

	t.Run("payment lookup edge cases", func(t *testing.T) {
		f := newPaymentFixture(t)
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(validPendingOrder(userID, orderID), nil).Times(3)

		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{}, nil).Once()
		_, err := f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		assert.ErrorIs(t, err, domainErrors.ErrRecordNotFound)

		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{{PaymentStatus: "paid", Amount: 500}}, nil).Once()
		_, err = f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)

		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{{PaymentStatus: "pending", Amount: 0}}, nil).Once()
		_, err = f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
	})

	t.Run("returns existing gateway order id without calling gateway", func(t *testing.T) {
		f := newPaymentFixture(t)
		existing := testGatewayOrderID
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(validPendingOrder(userID, orderID), nil).Once()
		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{
			{OrderID: orderID, PaymentStatus: "pending", Amount: 199.50, GatewayOrderID: &existing},
		}, nil).Once()

		got, err := f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		require.NoError(t, err)
		assert.Equal(t, testGatewayOrderID, got)
	})

	t.Run("gateway create and update errors", func(t *testing.T) {
		f := newPaymentFixture(t)
		gwErr := errors.New("razorpay unavailable")
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(validPendingOrder(userID, orderID), nil).Twice()
		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{
			{OrderID: orderID, PaymentStatus: "pending", Amount: 250.75},
		}, nil).Twice()

		f.gateway.EXPECT().CreateOrder(int64(25075), "ORD-1001").Return("", gwErr).Once()
		_, err := f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		assert.ErrorIs(t, err, gwErr)

		updateErr := errors.New("update failed")
		f.gateway.EXPECT().CreateOrder(int64(25075), "ORD-1001").Return(testGatewayOrderID, nil).Once()
		f.payRepo.EXPECT().Update(ctx, mock.MatchedBy(func(p *models.Payment) bool {
			return p.GatewayOrderID != nil && *p.GatewayOrderID == testGatewayOrderID
		})).Return(updateErr).Once()

		_, err = f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		assert.ErrorIs(t, err, updateErr)
	})

	t.Run("success", func(t *testing.T) {
		f := newPaymentFixture(t)
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(validPendingOrder(userID, orderID), nil).Once()
		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{
			{OrderID: orderID, PaymentStatus: "pending", Amount: 149.99},
		}, nil).Once()
		f.gateway.EXPECT().CreateOrder(int64(14999), "ORD-1001").Return(testGatewayOrderID, nil).Once()
		f.payRepo.EXPECT().Update(ctx, mock.MatchedBy(func(p *models.Payment) bool {
			return p.GatewayOrderID != nil && *p.GatewayOrderID == testGatewayOrderID
		})).Return(nil).Once()

		got, err := f.uc.CreateRazorpayOrder(ctx, userID, orderID)
		require.NoError(t, err)
		assert.Equal(t, testGatewayOrderID, got)
	})
}

func TestPaymentUseCase_VerifyRazorpayPayment(t *testing.T) {
	ctx := context.Background()
	userID := uint(11)
	orderID := uuid.New()
	gwOrder := testGatewayOrderID

	t.Run("input validation", func(t *testing.T) {
		f := newPaymentFixture(t)
		assert.ErrorIs(t, f.uc.VerifyRazorpayPayment(ctx, 0, orderID, gwOrder, testGatewayPayID, testRazorpaySig), domainErrors.ErrInvalidUserID)
		assert.ErrorIs(t, f.uc.VerifyRazorpayPayment(ctx, userID, uuid.Nil, gwOrder, testGatewayPayID, testRazorpaySig), domainErrors.ErrInvalidOrderID)
		assert.ErrorIs(t, f.uc.VerifyRazorpayPayment(ctx, userID, orderID, "", testGatewayPayID, testRazorpaySig), domainErrors.ErrInvalidInput)
		assert.ErrorIs(t, f.uc.VerifyRazorpayPayment(ctx, userID, orderID, gwOrder, "", testRazorpaySig), domainErrors.ErrInvalidInput)
		assert.ErrorIs(t, f.uc.VerifyRazorpayPayment(ctx, userID, orderID, gwOrder, testGatewayPayID, ""), domainErrors.ErrInvalidInput)
	})

	t.Run("signature and fetch validation failures", func(t *testing.T) {
		f := newPaymentFixture(t)
		order := validPendingOrder(userID, orderID)
		pay := models.Payment{OrderID: orderID, PaymentStatus: "pending", Amount: 100, GatewayOrderID: &gwOrder}

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(order, nil).Times(3)
		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{pay}, nil).Times(3)

		sigErr := errors.New("bad sig")
		f.gateway.EXPECT().VerifyPaymentSignature(gwOrder, testGatewayPayID, testRazorpaySig).Return(sigErr).Once()
		assert.ErrorIs(t, f.uc.VerifyRazorpayPayment(ctx, userID, orderID, gwOrder, testGatewayPayID, testRazorpaySig), sigErr)

		fetchErr := errors.New("fetch failed")
		f.gateway.EXPECT().VerifyPaymentSignature(gwOrder, testGatewayPayID, testRazorpaySig).Return(nil).Once()
		f.gateway.EXPECT().FetchPayment(testGatewayPayID).Return(nil, fetchErr).Once()
		assert.ErrorIs(t, f.uc.VerifyRazorpayPayment(ctx, userID, orderID, gwOrder, testGatewayPayID, testRazorpaySig), fetchErr)

		f.gateway.EXPECT().VerifyPaymentSignature(gwOrder, testGatewayPayID, testRazorpaySig).Return(nil).Once()
		f.gateway.EXPECT().FetchPayment(testGatewayPayID).Return(map[string]interface{}{
			"order_id": gwOrder,
			"currency": "USD",
			"status":   "captured",
			"amount":   float64(10000),
		}, nil).Once()
		assert.ErrorIs(t, f.uc.VerifyRazorpayPayment(ctx, userID, orderID, gwOrder, testGatewayPayID, testRazorpaySig), domainErrors.ErrInvalidInput)
	})

	t.Run("idempotent nil if locked order already paid", func(t *testing.T) {
		f := newPaymentFixture(t)
		order := validPendingOrder(userID, orderID)
		pay := models.Payment{OrderID: orderID, PaymentStatus: "pending", Amount: 100, GatewayOrderID: &gwOrder}

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(order, nil).Once()
		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{pay}, nil).Once()
		f.gateway.EXPECT().VerifyPaymentSignature(gwOrder, testGatewayPayID, testRazorpaySig).Return(nil).Once()
		f.gateway.EXPECT().FetchPayment(testGatewayPayID).Return(map[string]interface{}{
			"order_id": gwOrder,
			"currency": "INR",
			"status":   "captured",
			"amount":   float64(10000),
		}, nil).Once()

		f.expectTX(ctx)
		paidOrder := &models.Order{ID: orderID, UserID: userID, PaymentStatus: "paid"}
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(paidOrder, nil).Once()

		require.NoError(t, f.uc.VerifyRazorpayPayment(ctx, userID, orderID, gwOrder, testGatewayPayID, testRazorpaySig))
	})

	t.Run("tx errors when payment missing or mismatched gateway order", func(t *testing.T) {
		f := newPaymentFixture(t)
		order := validPendingOrder(userID, orderID)
		pay := models.Payment{OrderID: orderID, PaymentStatus: "pending", Amount: 100, GatewayOrderID: &gwOrder}
		wrongGW := "other_gw"

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(order, nil).Once()
		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{pay}, nil).Once()
		f.gateway.EXPECT().VerifyPaymentSignature(gwOrder, testGatewayPayID, testRazorpaySig).Return(nil).Once()
		f.gateway.EXPECT().FetchPayment(testGatewayPayID).Return(map[string]interface{}{
			"order_id": gwOrder,
			"currency": "INR",
			"status":   "captured",
			"amount":   float64(10000),
		}, nil).Once()

		f.expectTX(ctx)
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(validPendingOrder(userID, orderID), nil).Once()
		f.txPay.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{
			{OrderID: orderID, PaymentStatus: "pending", GatewayOrderID: &wrongGW},
		}, nil).Once()

		err := f.uc.VerifyRazorpayPayment(ctx, userID, orderID, gwOrder, testGatewayPayID, testRazorpaySig)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
	})

	t.Run("success marks payment paid and order confirmed", func(t *testing.T) {
		f := newPaymentFixture(t)
		order := validPendingOrder(userID, orderID)
		pay := models.Payment{OrderID: orderID, PaymentStatus: "pending", Amount: 250.50, GatewayOrderID: &gwOrder}

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(order, nil).Once()
		f.payRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{pay}, nil).Once()
		f.gateway.EXPECT().VerifyPaymentSignature(gwOrder, testGatewayPayID, testRazorpaySig).Return(nil).Once()
		f.gateway.EXPECT().FetchPayment(testGatewayPayID).Return(map[string]interface{}{
			"order_id": gwOrder,
			"currency": "INR",
			"status":   "captured",
			"amount":   float64(25050),
		}, nil).Once()

		f.expectTX(ctx)
		lockedOrder := validPendingOrder(userID, orderID)
		txPay := models.Payment{OrderID: orderID, PaymentStatus: "pending", Amount: 250.50, GatewayOrderID: &gwOrder}
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(lockedOrder, nil).Once()
		f.txPay.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{txPay}, nil).Once()
		f.txPay.EXPECT().Update(ctx, mock.MatchedBy(func(p *models.Payment) bool {
			return p.PaymentStatus == "paid" &&
				p.TransactionID != nil && *p.TransactionID == testGatewayPayID &&
				p.PaidAt != nil
		})).Return(nil).Once()
		f.txOrder.EXPECT().Update(ctx, mock.MatchedBy(func(o *models.Order) bool {
			return o.PaymentStatus == "paid" && o.OrderStatus == "confirmed"
		})).Return(nil).Once()

		require.NoError(t, f.uc.VerifyRazorpayPayment(ctx, userID, orderID, gwOrder, testGatewayPayID, testRazorpaySig))
	})
}

func TestPaymentUseCase_HandleRazorpayWebhook(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	gwOrder := testGatewayOrderID

	t.Run("signature and payload validation", func(t *testing.T) {
		f := newPaymentFixture(t)
		assert.ErrorIs(t, f.uc.HandleRazorpayWebhook(ctx, nil, "sig"), domainErrors.ErrInvalidInput)
		assert.ErrorIs(t, f.uc.HandleRazorpayWebhook(ctx, []byte(`{}`), ""), domainErrors.ErrInvalidInput)
		assert.ErrorIs(t, f.uc.HandleRazorpayWebhook(ctx, []byte(`{}`), "wrong_sig"), domainErrors.ErrInvalidInput)

		badJSON := []byte(`{bad`)
		assert.Error(t, f.uc.HandleRazorpayWebhook(ctx, badJSON, signWebhookPayload(badJSON)))
	})

	t.Run("ignores non-payment events", func(t *testing.T) {
		f := newPaymentFixture(t)
		payload := []byte(`{"event":"order.paid","payload":{"payment":{"entity":{"id":"p1","order_id":"o1","amount":100,"currency":"INR","status":"captured"}}}}`)
		require.NoError(t, f.uc.HandleRazorpayWebhook(ctx, payload, signWebhookPayload(payload)))
	})

	t.Run("rejects mismatched entity status or currency", func(t *testing.T) {
		f := newPaymentFixture(t)
		payload := []byte(`{"event":"payment.captured","payload":{"payment":{"entity":{"id":"p1","order_id":"o1","amount":100,"currency":"INR","status":"failed"}}}}`)
		assert.ErrorIs(t, f.uc.HandleRazorpayWebhook(ctx, payload, signWebhookPayload(payload)), domainErrors.ErrInvalidInput)
	})

	t.Run("payment.captured idempotent when already paid", func(t *testing.T) {
		f := newPaymentFixture(t)
		payload := []byte(`{"event":"payment.captured","payload":{"payment":{"entity":{"id":"` + testGatewayPayID + `","order_id":"` + gwOrder + `","amount":10000,"currency":"INR","status":"captured"}}}}`)

		f.expectTX(ctx)
		f.txPay.EXPECT().GetByGatewayOrderID(ctx, gwOrder).Return(&models.Payment{
			OrderID: orderID, Amount: 100, GatewayOrderID: &gwOrder, PaymentStatus: "paid",
		}, nil).Once()
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(&models.Order{
			ID: orderID, PaymentStatus: "paid",
		}, nil).Once()
		f.txPay.EXPECT().GetByGatewayOrderIDForUpdate(ctx, gwOrder).Return(&models.Payment{
			OrderID: orderID, Amount: 100, GatewayOrderID: &gwOrder, PaymentStatus: "paid",
		}, nil).Once()

		require.NoError(t, f.uc.HandleRazorpayWebhook(ctx, payload, signWebhookPayload(payload)))
	})

	t.Run("payment.captured updates payment and order", func(t *testing.T) {
		f := newPaymentFixture(t)
		payload := []byte(`{"event":"payment.captured","payload":{"payment":{"entity":{"id":"` + testGatewayPayID + `","order_id":"` + gwOrder + `","amount":50000,"currency":"INR","status":"captured"}}}}`)

		f.expectTX(ctx)
		initialPay := &models.Payment{OrderID: orderID, Amount: 500, GatewayOrderID: &gwOrder, PaymentStatus: "pending"}
		lockedOrder := &models.Order{ID: orderID, OrderStatus: "pending", PaymentStatus: "pending"}
		txPay := &models.Payment{OrderID: orderID, Amount: 500, GatewayOrderID: &gwOrder, PaymentStatus: "pending"}

		f.txPay.EXPECT().GetByGatewayOrderID(ctx, gwOrder).Return(initialPay, nil).Once()
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(lockedOrder, nil).Once()
		f.txPay.EXPECT().GetByGatewayOrderIDForUpdate(ctx, gwOrder).Return(txPay, nil).Once()
		f.txPay.EXPECT().Update(ctx, mock.MatchedBy(func(p *models.Payment) bool {
			return p.PaymentStatus == "paid" &&
				p.TransactionID != nil && *p.TransactionID == testGatewayPayID &&
				p.PaidAt != nil
		})).Return(nil).Once()
		f.txOrder.EXPECT().Update(ctx, mock.MatchedBy(func(o *models.Order) bool {
			return o.PaymentStatus == "paid" && o.OrderStatus == "confirmed"
		})).Return(nil).Once()

		require.NoError(t, f.uc.HandleRazorpayWebhook(ctx, payload, signWebhookPayload(payload)))
	})

	t.Run("payment.failed updates payment failure reason and order status", func(t *testing.T) {
		f := newPaymentFixture(t)
		payload := []byte(`{"event":"payment.failed","payload":{"payment":{"entity":{"id":"` + testGatewayPayID + `","order_id":"` + gwOrder + `","amount":50000,"currency":"INR","status":"failed","error_description":"insufficient funds"}}}}`)

		f.expectTX(ctx)
		initialPay := &models.Payment{OrderID: orderID, Amount: 500, GatewayOrderID: &gwOrder, PaymentStatus: "pending"}
		lockedOrder := &models.Order{ID: orderID, OrderStatus: "pending", PaymentStatus: "pending"}
		txPay := &models.Payment{OrderID: orderID, Amount: 500, GatewayOrderID: &gwOrder, PaymentStatus: "pending"}

		f.txPay.EXPECT().GetByGatewayOrderID(ctx, gwOrder).Return(initialPay, nil).Once()
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(lockedOrder, nil).Once()
		f.txPay.EXPECT().GetByGatewayOrderIDForUpdate(ctx, gwOrder).Return(txPay, nil).Once()
		f.txPay.EXPECT().Update(ctx, mock.MatchedBy(func(p *models.Payment) bool {
			return p.PaymentStatus == "failed" &&
				p.FailureReason != nil && *p.FailureReason == "insufficient funds"
		})).Return(nil).Once()
		f.txOrder.EXPECT().Update(ctx, mock.MatchedBy(func(o *models.Order) bool {
			return o.PaymentStatus == "failed" && o.OrderStatus == "confirmed"
		})).Return(nil).Once()

		require.NoError(t, f.uc.HandleRazorpayWebhook(ctx, payload, signWebhookPayload(payload)))
	})
}
