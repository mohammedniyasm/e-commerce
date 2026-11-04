package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *models.Wallet) error
	GetByUserID(ctx context.Context, userID uint) (*models.Wallet, error)
	GetByUserIDForUpdate(ctx context.Context, userID uint) (*models.Wallet, error)
	Update(ctx context.Context, wallet *models.Wallet) error

	CreateTransaction(ctx context.Context, transaction *models.WalletTransaction) error
	// ApplyTransaction(ctx context.Context, transaction *models.WalletTransaction) error
	GetTransactionByID(ctx context.Context, transactionID uint, walletID uint) (*models.WalletTransaction, error)
	ListTransactions(ctx context.Context, walletID uint, limit int, page int, sort *string) ([]models.WalletTransaction, int64, error)
}
