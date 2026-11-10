package wallet

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"io"
	"log/slog"
	"math"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupWalletTest(t *testing.T) (
	context.Context,
	*mocks.MockWalletRepository,
	*mocks.MockWalletTopUpRepository,
	*mocks.MockPaymentGateway,
	*mocks.MockUnitOfWork,
	*WalletUseCase,
) {
	ctx := context.Background()
	walletRepo := mocks.NewMockWalletRepository(t)
	topUpRepo := mocks.NewMockWalletTopUpRepository(t)
	paymentGateway := mocks.NewMockPaymentGateway(t)
	uow := mocks.NewMockUnitOfWork(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	uc := NewWalletUseCase(walletRepo, topUpRepo, paymentGateway, uow, "rzp_test_key", logger)
	return ctx, walletRepo, topUpRepo, paymentGateway, uow, uc
}

func TestGetWallet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, walletRepo, _, _, _, uc := setupWalletTest(t)
		expectedWallet := &models.Wallet{
			ID:      10,
			Balance: 500.0,
		}

		walletRepo.EXPECT().GetByUserID(ctx, uint(10)).Return(expectedWallet, nil).Once()

		wallet, err := uc.GetWallet(ctx, 10)
		require.NoError(t, err)
		require.Equal(t, expectedWallet, wallet)
	})

	t.Run("Invalid UserID", func(t *testing.T) {
		ctx, _, _, _, _, uc := setupWalletTest(t)

		wallet, err := uc.GetWallet(ctx, 0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		require.Nil(t, wallet)
	})

	t.Run("Repo Error", func(t *testing.T) {
		ctx, walletRepo, _, _, _, uc := setupWalletTest(t)
		dbErr := errors.New("db connection failure")

		walletRepo.EXPECT().GetByUserID(ctx, uint(10)).Return(nil, dbErr).Once()

		wallet, err := uc.GetWallet(ctx, 10)
		require.ErrorIs(t, err, dbErr)
		require.Nil(t, wallet)
	})
}

func TestDebitWallet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, _, _, _, uow, uc := setupWalletTest(t)
		txWalletRepo := mocks.NewMockWalletRepository(t)

		currentWallet := &models.Wallet{
			ID:      10,
			Balance: 300.0,
		}

		uow.EXPECT().WithinTransaction(ctx, mock.Anything).RunAndReturn(
			func(ctx context.Context, fn func(repos interfaces.TransactionRepositories) error) error {
				return fn(interfaces.TransactionRepositories{
					Wallet: txWalletRepo,
				})
			},
		).Once()

		txWalletRepo.EXPECT().GetByUserIDForUpdate(ctx, uint(10)).Return(currentWallet, nil).Once()
		txWalletRepo.EXPECT().Update(ctx, mock.MatchedBy(func(w *models.Wallet) bool {
			return w.Balance == 200.0
		})).Return(nil).Once()
		txWalletRepo.EXPECT().CreateTransaction(ctx, mock.MatchedBy(func(tx *models.WalletTransaction) bool {
			return tx.WalletID == 10 &&
				tx.Amount == 100.0 &&
				tx.BalanceBefore == 300.0 &&
				tx.BalanceAfter == 200.0 &&
				tx.TransactionType == models.WalletTransactionDebit
		})).Return(nil).Once()

		err := uc.DebitWallet(ctx, 10, 100.0, models.WalletSourceOrderPayment, nil, nil)
		require.NoError(t, err)
	})

	t.Run("Invalid UserID", func(t *testing.T) {
		ctx, _, _, _, _, uc := setupWalletTest(t)

		err := uc.DebitWallet(ctx, 0, 100.0, models.WalletSourceOrderPayment, nil, nil)
		require.ErrorIs(t, err, domainerrors.ErrInvalidUserID)
	})

	t.Run("Invalid Amount", func(t *testing.T) {
		ctx, _, _, _, _, uc := setupWalletTest(t)

		err := uc.DebitWallet(ctx, 10, -50.0, models.WalletSourceOrderPayment, nil, nil)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	t.Run("Insufficient Balance", func(t *testing.T) {
		ctx, _, _, _, uow, uc := setupWalletTest(t)
		txWalletRepo := mocks.NewMockWalletRepository(t)

		currentWallet := &models.Wallet{
			ID:      10,
			Balance: 50.0,
		}

		uow.EXPECT().WithinTransaction(ctx, mock.Anything).RunAndReturn(
			func(ctx context.Context, fn func(repos interfaces.TransactionRepositories) error) error {
				return fn(interfaces.TransactionRepositories{
					Wallet: txWalletRepo,
				})
			},
		).Once()

		txWalletRepo.EXPECT().GetByUserIDForUpdate(ctx, uint(10)).Return(currentWallet, nil).Once()

		err := uc.DebitWallet(ctx, 10, 100.0, models.WalletSourceOrderPayment, nil, nil)
		require.ErrorIs(t, err, domainerrors.ErrInsufficientBalance)
	})
}

func TestCreditWallet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, _, _, _, uow, uc := setupWalletTest(t)
		txWalletRepo := mocks.NewMockWalletRepository(t)

		currentWallet := &models.Wallet{
			ID:      10,
			Balance: 150.0,
		}

		uow.EXPECT().WithinTransaction(ctx, mock.Anything).RunAndReturn(
			func(ctx context.Context, fn func(repos interfaces.TransactionRepositories) error) error {
				return fn(interfaces.TransactionRepositories{
					Wallet: txWalletRepo,
				})
			},
		).Once()

		txWalletRepo.EXPECT().GetByUserIDForUpdate(ctx, uint(10)).Return(currentWallet, nil).Once()
		txWalletRepo.EXPECT().Update(ctx, mock.MatchedBy(func(w *models.Wallet) bool {
			return w.Balance == 250.0
		})).Return(nil).Once()
		txWalletRepo.EXPECT().CreateTransaction(ctx, mock.MatchedBy(func(tx *models.WalletTransaction) bool {
			return tx.WalletID == 10 &&
				tx.Amount == 100.0 &&
				tx.BalanceBefore == 150.0 &&
				tx.BalanceAfter == 250.0 &&
				tx.TransactionType == models.WalletTransactionCredit
		})).Return(nil).Once()

		err := uc.CreditWallet(ctx, 10, 100.0, models.WalletSourceRefund, nil, nil)
		require.NoError(t, err)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		ctx, _, _, _, _, uc := setupWalletTest(t)

		err := uc.CreditWallet(ctx, 0, 100.0, models.WalletSourceRefund, nil, nil)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		err = uc.CreditWallet(ctx, 10, 0, models.WalletSourceRefund, nil, nil)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})
}

func TestGetTransactionByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, walletRepo, _, _, _, uc := setupWalletTest(t)
		expectedTx := &models.WalletTransaction{
			ID:       5,
			WalletID: 10,
			Amount:   200.0,
		}

		walletRepo.EXPECT().GetTransactionByID(ctx, uint(5), uint(10)).Return(expectedTx, nil).Once()

		tx, err := uc.GetTransactionByID(ctx, 10, 5)
		require.NoError(t, err)
		require.Equal(t, expectedTx, tx)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		ctx, _, _, _, _, uc := setupWalletTest(t)

		tx, err := uc.GetTransactionByID(ctx, 0, 5)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		require.Nil(t, tx)
	})
}

func TestListTransactions(t *testing.T) {
	t.Run("Success With Normalization", func(t *testing.T) {
		ctx, walletRepo, _, _, _, uc := setupWalletTest(t)
		expectedTxs := []models.WalletTransaction{
			{ID: 1, WalletID: 10, Amount: 100},
			{ID: 2, WalletID: 10, Amount: 200},
		}

		// limit 0 should be normalized to 10, page 0 normalized to 1
		walletRepo.EXPECT().ListTransactions(ctx, uint(10), 10, 1, (*string)(nil)).Return(expectedTxs, int64(2), nil).Once()

		txs, total, err := uc.ListTransactions(ctx, 10, 0, 0, nil)
		require.NoError(t, err)
		require.Equal(t, int64(2), total)
		require.Len(t, txs, 2)
	})

	t.Run("Max Limit Capped", func(t *testing.T) {
		ctx, walletRepo, _, _, _, uc := setupWalletTest(t)

		// limit 150 should be capped to 100
		walletRepo.EXPECT().ListTransactions(ctx, uint(10), 100, 2, (*string)(nil)).Return(nil, int64(0), nil).Once()

		_, _, err := uc.ListTransactions(ctx, 10, 150, 2, nil)
		require.NoError(t, err)
	})

	t.Run("Invalid UserID", func(t *testing.T) {
		ctx, _, _, _, _, uc := setupWalletTest(t)

		_, _, err := uc.ListTransactions(ctx, 0, 10, 1, nil)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})
}

func TestCreateTopUpOrder(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, walletRepo, topUpRepo, paymentGateway, _, uc := setupWalletTest(t)

		walletRepo.EXPECT().GetByUserID(ctx, uint(10)).Return(&models.Wallet{ID: 10}, nil).Once()
		paymentGateway.EXPECT().CreateOrder(int64(50000), mock.MatchedBy(func(rcpt string) bool {
			return len(rcpt) > 4
		})).Return("order_rzp_123", nil).Once()

		topUpRepo.EXPECT().Create(ctx, mock.MatchedBy(func(tu *models.WalletTopUp) bool {
			return tu.WalletID == 10 &&
				tu.Amount == 500.0 &&
				tu.GatewayOrderID == "order_rzp_123" &&
				tu.Status == models.WalletTopUpPending
		})).Return(nil).Once()

		orderID, err := uc.CreateTopUpOrder(ctx, 10, 500.0)
		require.NoError(t, err)
		require.Equal(t, "order_rzp_123", orderID)
	})

	t.Run("Invalid Input Values", func(t *testing.T) {
		ctx, _, _, _, _, uc := setupWalletTest(t)

		_, err := uc.CreateTopUpOrder(ctx, 0, 100.0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, err = uc.CreateTopUpOrder(ctx, 10, -10.0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, err = uc.CreateTopUpOrder(ctx, 10, math.NaN())
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, err = uc.CreateTopUpOrder(ctx, 10, math.Inf(1))
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	t.Run("Payment Gateway Error", func(t *testing.T) {
		ctx, walletRepo, _, paymentGateway, _, uc := setupWalletTest(t)

		walletRepo.EXPECT().GetByUserID(ctx, uint(10)).Return(&models.Wallet{ID: 10}, nil).Once()
		gwErr := errors.New("gateway unavailable")
		paymentGateway.EXPECT().CreateOrder(int64(10000), mock.Anything).Return("", gwErr).Once()

		_, err := uc.CreateTopUpOrder(ctx, 10, 100.0)
		require.ErrorIs(t, err, gwErr)
	})
}

func TestVerifyTopUpPayment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, _, _, paymentGateway, uow, uc := setupWalletTest(t)
		txWalletRepo := mocks.NewMockWalletRepository(t)
		txTopUpRepo := mocks.NewMockWalletTopUpRepository(t)

		orderID := "order_rzp_123"
		paymentID := "pay_rzp_456"
		sig := "signature_valid"

		paymentGateway.EXPECT().VerifyPaymentSignature(orderID, paymentID, sig).Return(nil).Once()
		paymentGateway.EXPECT().FetchPayment(paymentID).Return(map[string]interface{}{
			"order_id": orderID,
			"currency": "INR",
			"status":   "captured",
			"amount":   float64(50000), // 500 INR in paise
		}, nil).Once()

		uow.EXPECT().WithinTransaction(ctx, mock.Anything).RunAndReturn(
			func(ctx context.Context, fn func(repos interfaces.TransactionRepositories) error) error {
				return fn(interfaces.TransactionRepositories{
					Wallet:      txWalletRepo,
					WalletTopUp: txTopUpRepo,
				})
			},
		).Once()

		topUp := &models.WalletTopUp{
			ID:             1,
			WalletID:       10,
			Amount:         500.0,
			GatewayOrderID: orderID,
			Status:         models.WalletTopUpPending,
		}
		txTopUpRepo.EXPECT().GetByGatewayOrderIDForUpdate(ctx, orderID).Return(topUp, nil).Once()

		wallet := &models.Wallet{
			ID:      10,
			Balance: 100.0,
		}
		txWalletRepo.EXPECT().GetByUserIDForUpdate(ctx, uint(10)).Return(wallet, nil).Once()
		txWalletRepo.EXPECT().Update(ctx, mock.MatchedBy(func(w *models.Wallet) bool {
			return w.Balance == 600.0
		})).Return(nil).Once()

		txWalletRepo.EXPECT().CreateTransaction(ctx, mock.MatchedBy(func(tx *models.WalletTransaction) bool {
			return tx.WalletID == 10 &&
				tx.Amount == 500.0 &&
				tx.BalanceBefore == 100.0 &&
				tx.BalanceAfter == 600.0 &&
				tx.TransactionType == models.WalletTransactionCredit &&
				tx.Source == models.WalletSourceTopUp
		})).Return(nil).Once()

		txTopUpRepo.EXPECT().Update(ctx, mock.MatchedBy(func(tu *models.WalletTopUp) bool {
			return tu.Status == models.WalletTopUpPaid &&
				tu.GatewayPaymentID != nil &&
				*tu.GatewayPaymentID == paymentID
		})).Return(nil).Once()

		err := uc.VerifyTopUpPayment(ctx, 10, orderID, paymentID, sig)
		require.NoError(t, err)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		ctx, _, _, _, _, uc := setupWalletTest(t)

		err := uc.VerifyTopUpPayment(ctx, 0, "order_1", "pay_1", "sig_1")
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		err = uc.VerifyTopUpPayment(ctx, 10, "", "pay_1", "sig_1")
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	t.Run("Signature Verification Failed", func(t *testing.T) {
		ctx, _, _, paymentGateway, _, uc := setupWalletTest(t)

		sigErr := errors.New("invalid signature")
		paymentGateway.EXPECT().VerifyPaymentSignature("order_1", "pay_1", "invalid_sig").Return(sigErr).Once()

		err := uc.VerifyTopUpPayment(ctx, 10, "order_1", "pay_1", "invalid_sig")
		require.ErrorIs(t, err, sigErr)
	})

	t.Run("Invalid Currency In Gateway", func(t *testing.T) {
		ctx, _, _, paymentGateway, _, uc := setupWalletTest(t)

		paymentGateway.EXPECT().VerifyPaymentSignature("order_1", "pay_1", "sig_1").Return(nil).Once()
		paymentGateway.EXPECT().FetchPayment("pay_1").Return(map[string]interface{}{
			"order_id": "order_1",
			"currency": "USD",
			"status":   "captured",
			"amount":   float64(50000),
		}, nil).Once()

		err := uc.VerifyTopUpPayment(ctx, 10, "order_1", "pay_1", "sig_1")
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	t.Run("Payment Not Captured", func(t *testing.T) {
		ctx, _, _, paymentGateway, _, uc := setupWalletTest(t)

		paymentGateway.EXPECT().VerifyPaymentSignature("order_1", "pay_1", "sig_1").Return(nil).Once()
		paymentGateway.EXPECT().FetchPayment("pay_1").Return(map[string]interface{}{
			"order_id": "order_1",
			"currency": "INR",
			"status":   "failed",
			"amount":   float64(50000),
		}, nil).Once()

		err := uc.VerifyTopUpPayment(ctx, 10, "order_1", "pay_1", "sig_1")
		require.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})
}
