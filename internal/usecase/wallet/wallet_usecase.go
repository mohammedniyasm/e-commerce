package wallet

import (
	"context"
	"log/slog"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
)

type WalletUseCase struct {
	walletRepo interfaces.WalletRepository
	unitOfWork interfaces.UnitOfWork
	log        *slog.Logger
}

func NewWalletUseCase(
	walletRepo interfaces.WalletRepository,
	unitOfWork interfaces.UnitOfWork,
	log *slog.Logger,
) *WalletUseCase {
	return &WalletUseCase{
		walletRepo: walletRepo,
		unitOfWork: unitOfWork,
		log:        log,
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

var _ interfaces.WalletUseCase = (*WalletUseCase)(nil)
