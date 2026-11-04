package wallet

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
)

type WalletUseCase struct {
	walletRepo      interfaces.WalletRepository
	walletTopUpRepo interfaces.WalletTopUpRepository
	paymentGateway  interfaces.PaymentGateway
	unitOfWork      interfaces.UnitOfWork
	razorpayKeyID   string
	log             *slog.Logger
}

func NewWalletUseCase(
	walletRepo interfaces.WalletRepository,
	walletTopUpRepo interfaces.WalletTopUpRepository,
	paymentGateway interfaces.PaymentGateway,
	unitOfWork interfaces.UnitOfWork,
	razorpayKeyID string,
	log *slog.Logger,
) *WalletUseCase {
	return &WalletUseCase{
		walletRepo:      walletRepo,
		walletTopUpRepo: walletTopUpRepo,
		paymentGateway:  paymentGateway,
		unitOfWork:      unitOfWork,
		razorpayKeyID:   razorpayKeyID,
		log:             log,
	}
}
func (u *WalletUseCase) GetWallet(ctx context.Context, userID uint) (*models.Wallet, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	wallet, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to retrieve user wallet",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}
	u.log.InfoContext(
		ctx,
		"user wallet retrieved",
		"user_id", userID,
	)
	return wallet, nil
}
func (u *WalletUseCase) DebitWallet(ctx context.Context, userID uint, amount float64, source models.WalletTransactionSource, referenceID *uuid.UUID, description *string) error {
	if userID <= 0 {
		return domainerrors.ErrInvalidUserID
	}
	if amount <= 0 {
		return domainerrors.ErrInvalidInput
	}
	u.log.InfoContext(ctx, "wallet debit initiated",
		"user_id", userID,
		"amount", amount,
		"source", source,
	)
	err := u.unitOfWork.WithinTransaction(ctx, func(repos interfaces.TransactionRepositories) error {
		wallet, err := repos.Wallet.GetByUserIDForUpdate(ctx, userID)
		if err != nil {
			return err
		}
		if wallet.Balance < amount {
			return domainerrors.ErrInsufficientBalance
		}
		balanceBefore := wallet.Balance
		balanceAfter := balanceBefore - amount
		wallet.Balance = balanceAfter
		if err := repos.Wallet.Update(ctx, wallet); err != nil {
			return err
		}
		transaction := &models.WalletTransaction{
			WalletID:        userID,
			TransactionType: models.WalletTransactionDebit,
			Source:          source,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceAfter,
			ReferenceID:     referenceID,
			Description:     description,
		}
		if err := repos.Wallet.CreateTransaction(ctx, transaction); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		u.log.ErrorContext(ctx, "wallet debit failed",
			"user_id", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}

	u.log.InfoContext(ctx, "wallet debited successfully",
		"user_id", userID,
		"amount", amount,
	)
	return nil
}
func (u *WalletUseCase) CreditWallet(ctx context.Context, userID uint, amount float64, source models.WalletTransactionSource, referenceID *uuid.UUID, description *string) error {
	if userID <= 0 || amount <= 0 {
		return domainerrors.ErrInvalidInput
	}
	u.log.InfoContext(ctx, "wallet credit initiated",
		"user_id", userID,
		"amount", amount,
		"source", source,
	)
	err := u.unitOfWork.WithinTransaction(ctx, func(repos interfaces.TransactionRepositories) error {
		wallet, err := repos.Wallet.GetByUserIDForUpdate(ctx, userID)
		if err != nil {
			return err
		}
		balanceBefore := wallet.Balance
		balanceAfter := balanceBefore + amount
		wallet.Balance = balanceAfter
		if err := repos.Wallet.Update(ctx, wallet); err != nil {
			return err
		}
		transaction := &models.WalletTransaction{
			WalletID:        userID,
			TransactionType: models.WalletTransactionCredit,
			Source:          source,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceAfter,
			ReferenceID:     referenceID,
			Description:     description,
		}
		if err := repos.Wallet.CreateTransaction(ctx, transaction); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		u.log.ErrorContext(ctx, "wallet credit failed",
			"user_id", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}
	u.log.InfoContext(ctx, "wallet credited successfully",
		"user_id", userID,
		"amount", amount,
	)
	return nil
}
func (u *WalletUseCase) GetTransactionByID(ctx context.Context, userID uint, transactionID uint) (*models.WalletTransaction, error) {
	if userID <= 0 || transactionID <= 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	transaction, err := u.walletRepo.GetTransactionByID(
		ctx,
		transactionID,
		userID,
	)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to retrieve wallet transaction",
			"user_id", userID,
			"transaction_id", transactionID,
			"error", err,
		)
		return nil, err
	}
	return transaction, nil
}
func (u *WalletUseCase) ListTransactions(ctx context.Context, userID uint, limit int, page int, sort *string) ([]models.WalletTransaction, int64, error) {
	if userID == 0 {
		return nil, 0, domainerrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	transactions, total, err := u.walletRepo.ListTransactions(
		ctx,
		userID,
		limit,
		page,
		sort,
	)
	if err != nil {
		u.log.ErrorContext(ctx, "failed to list wallet transactions",
			"user_id", userID,
			"error", err,
		)
		return nil, 0, err
	}
	return transactions, total, nil
}
func (u *WalletUseCase) CreateTopUpOrder(ctx context.Context, userID uint, amount float64) (string, error) {
	if userID == 0 || amount <= 0 ||
		math.IsNaN(amount) || math.IsInf(amount, 0) {
		return "", domainerrors.ErrInvalidInput
	}
	amountInPaise := int64(math.Round(amount * 100))
	if amountInPaise <= 0 {
		return "", domainerrors.ErrInvalidInput
	}
	_, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	receipt := "wtu_" + uuid.NewString()
	gatewayOrderID, err := u.paymentGateway.CreateOrder(
		amountInPaise,
		receipt,
	)
	if err != nil {
		u.log.ErrorContext(ctx,
			"failed to create wallet top-up razorpay order",
			"user_id", userID,
			"error", err,
		)
		return "", err
	}
	now := time.Now()
	topUp := &models.WalletTopUp{
		WalletID:       userID,
		Amount:         float64(amountInPaise) / 100,
		GatewayOrderID: gatewayOrderID,
		Status:         models.WalletTopUpPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := u.walletTopUpRepo.Create(ctx, topUp); err != nil {
		u.log.ErrorContext(ctx,
			"failed to save wallet top-up",
			"user_id", userID,
			"gateway_order_id", gatewayOrderID,
			"error", err,
		)
		return "", err
	}
	u.log.InfoContext(ctx,
		"wallet top-up order created",
		"user_id", userID,
		"amount", topUp.Amount,
		"gateway_order_id", gatewayOrderID,
	)
	return gatewayOrderID, nil
}
func (u *WalletUseCase) VerifyTopUpPayment(ctx context.Context, userID uint, razorpayOrderID string, razorpayPaymentID string, signature string) error {
	if userID == 0 ||
		razorpayOrderID == "" ||
		razorpayPaymentID == "" ||
		signature == "" {
		return domainerrors.ErrInvalidInput
	}
	if err := u.paymentGateway.VerifyPaymentSignature(
		razorpayOrderID,
		razorpayPaymentID,
		signature,
	); err != nil {
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
	gatewayAmount, ok := gatewayPayment["amount"].(float64)
	if !ok {
		return domainerrors.ErrInvalidInput
	}
	now := time.Now()
	return u.unitOfWork.WithinTransaction(
		ctx,
		func(repos interfaces.TransactionRepositories) error {
			topUp, err := repos.WalletTopUp.GetByGatewayOrderIDForUpdate(
				ctx,
				razorpayOrderID,
			)
			if err != nil {
				return err
			}
			if topUp.WalletID != userID {
				return domainerrors.ErrRecordNotFound
			}
			expectedAmount := int64(math.Round(topUp.Amount * 100))
			if int64(gatewayAmount) != expectedAmount {
				return domainerrors.ErrInvalidInput
			}
			if topUp.Status == models.WalletTopUpPaid {
				if topUp.GatewayPaymentID != nil &&
					*topUp.GatewayPaymentID == razorpayPaymentID {
					return nil
				}
				return domainerrors.ErrInvalidInput
			}
			if topUp.Status != models.WalletTopUpPending {
				return domainerrors.ErrInvalidInput
			}
			wallet, err := repos.Wallet.GetByUserIDForUpdate(
				ctx,
				topUp.WalletID,
			)
			if err != nil {
				return err
			}
			balanceBefore := wallet.Balance
			balanceAfter := balanceBefore + topUp.Amount
			wallet.Balance = balanceAfter
			wallet.UpdatedAt = now
			if err := repos.Wallet.Update(ctx, wallet); err != nil {
				return err
			}
			description := fmt.Sprintf(
				"Wallet top-up through Razorpay. Payment ID: %s",
				razorpayPaymentID,
			)
			transaction := &models.WalletTransaction{
				WalletID:        wallet.ID,
				TransactionType: models.WalletTransactionCredit,
				Source:          models.WalletSourceTopUp,
				Amount:          topUp.Amount,
				BalanceBefore:   balanceBefore,
				BalanceAfter:    balanceAfter,
				Description:     &description,
				CreatedAt:       now,
			}
			if err := repos.Wallet.CreateTransaction(
				ctx,
				transaction,
			); err != nil {
				return err
			}
			topUp.GatewayPaymentID = &razorpayPaymentID
			topUp.Status = models.WalletTopUpPaid
			topUp.PaidAt = &now
			topUp.UpdatedAt = now
			if err := repos.WalletTopUp.Update(ctx, topUp); err != nil {
				return err
			}
			return nil
		},
	)
}
func (u *WalletUseCase) ProcessTopUpWebhook(ctx context.Context, razorpayOrderID string, razorpayPaymentID string, amount int64, currency string) error {
	if razorpayOrderID == "" ||
		razorpayPaymentID == "" ||
		amount <= 0 ||
		currency != "INR" {
		return domainerrors.ErrInvalidInput
	}
	now := time.Now()
	return u.unitOfWork.WithinTransaction(
		ctx,
		func(repos interfaces.TransactionRepositories) error {
			topUp, err := repos.WalletTopUp.GetByGatewayOrderIDForUpdate(
				ctx,
				razorpayOrderID,
			)
			if err != nil {
				return err
			}
			expectedAmount := int64(math.Round(topUp.Amount * 100))
			if expectedAmount != amount {
				return domainerrors.ErrInvalidInput
			}
			if topUp.Status == models.WalletTopUpPaid {
				if topUp.GatewayPaymentID != nil &&
					*topUp.GatewayPaymentID == razorpayPaymentID {
					return nil
				}
				return domainerrors.ErrInvalidInput
			}
			if topUp.Status != models.WalletTopUpPending {
				return domainerrors.ErrInvalidInput
			}
			wallet, err := repos.Wallet.GetByUserIDForUpdate(
				ctx,
				topUp.WalletID,
			)
			if err != nil {
				return err
			}
			balanceBefore := wallet.Balance
			balanceAfter := balanceBefore + topUp.Amount
			wallet.Balance = balanceAfter
			wallet.UpdatedAt = now
			if err := repos.Wallet.Update(ctx, wallet); err != nil {
				return err
			}
			description := fmt.Sprintf(
				"Wallet top-up through Razorpay. Order ID: %s, Payment ID: %s",
				razorpayOrderID,
				razorpayPaymentID,
			)
			transaction := &models.WalletTransaction{
				WalletID:        wallet.ID,
				TransactionType: models.WalletTransactionCredit,
				Source:          models.WalletSourceTopUp,
				Amount:          topUp.Amount,
				BalanceBefore:   balanceBefore,
				BalanceAfter:    balanceAfter,
				ReferenceID:     nil,
				Description:     &description,
				CreatedAt:       now,
			}
			if err := repos.Wallet.CreateTransaction(
				ctx,
				transaction,
			); err != nil {
				return err
			}
			topUp.GatewayPaymentID = &razorpayPaymentID
			topUp.Status = models.WalletTopUpPaid
			topUp.PaidAt = &now
			topUp.UpdatedAt = now
			if err := repos.WalletTopUp.Update(
				ctx,
				topUp,
			); err != nil {
				return err
			}
			return nil
		},
	)
}

var _ interfaces.WalletUseCase = (*WalletUseCase)(nil)
