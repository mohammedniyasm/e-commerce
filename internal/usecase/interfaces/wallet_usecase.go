package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

type WalletUseCase interface {
	GetWallet(ctx context.Context, userID uint) (*models.Wallet, error)
	CreditWallet(ctx context.Context, userID uint, amount float64, source models.WalletTransactionSource, referenceID *uuid.UUID, description *string) error
	DebitWallet(ctx context.Context, userID uint, amount float64, source models.WalletTransactionSource, referenceID *uuid.UUID, description *string) error
	GetTransactionByID(ctx context.Context, userID uint, transactionID uint) (*models.WalletTransaction, error)
	ListTransactions(ctx context.Context, userID uint, limit int, page int, sort *string) ([]models.WalletTransaction, int64, error)
}
