package response

import (
	"ecommerce/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type WalletResponse struct {
	UserID    uint      `json:"user_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type WalletTransactionResponse struct {
	ID              uint       `json:"id"`
	TransactionType string     `json:"transaction_type"`
	Source          string     `json:"source"`
	Amount          float64    `json:"amount"`
	BalanceBefore   float64    `json:"balance_before"`
	BalanceAfter    float64    `json:"balance_after"`
	ReferenceID     *uuid.UUID `json:"reference_id"`
	Description     *string    `json:"description"`
	CreatedAt       time.Time  `json:"created_at"`
}
type WalletTransactionListResponse struct {
	ID              uint       `json:"id"`
	TransactionType string     `json:"transaction_type"`
	Source          string     `json:"source"`
	Amount          float64    `json:"amount"`
	ReferenceID     *uuid.UUID `json:"reference_id"`
	Description     *string    `json:"description"`
	CreatedAt       time.Time  `json:"created_at"`
}

func ToWalletResponse(wallet *models.Wallet) WalletResponse {
	return WalletResponse{
		UserID:    wallet.ID,
		Balance:   wallet.Balance,
		CreatedAt: wallet.CreatedAt,
		UpdatedAt: wallet.UpdatedAt,
	}
}
func ToWalletTransactionResponse(transaction *models.WalletTransaction) WalletTransactionResponse {
	return WalletTransactionResponse{
		ID:              transaction.ID,
		TransactionType: string(transaction.TransactionType),
		Source:          string(transaction.Source),
		Amount:          transaction.Amount,
		BalanceBefore:   transaction.BalanceBefore,
		BalanceAfter:    transaction.BalanceAfter,
		ReferenceID:     transaction.ReferenceID,
		Description:     transaction.Description,
		CreatedAt:       transaction.CreatedAt,
	}
}
func ToWalletTransactionListResponse(transactions []models.WalletTransaction) []WalletTransactionListResponse {
	result := make([]WalletTransactionListResponse, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, WalletTransactionListResponse{
			ID:              transaction.ID,
			TransactionType: string(transaction.TransactionType),
			Source:          string(transaction.Source),
			Amount:          transaction.Amount,
			ReferenceID:     transaction.ReferenceID,
			Description:     transaction.Description,
			CreatedAt:       transaction.CreatedAt,
		})
	}
	return result
}
