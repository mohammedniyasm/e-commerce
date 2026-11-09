package interfaces

import (
	"context"
	// "ecommerce/internal/usecase/interfaces"
)

type UnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(repos TransactionRepositories) error) error
}
