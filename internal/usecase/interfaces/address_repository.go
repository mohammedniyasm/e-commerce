package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type AddressRepository interface {
	Create(ctx context.Context, address *models.Address) error
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	ClearDefaultByUserID(ctx context.Context, userID uint) error
	GetByUserID(ctx context.Context, userID uint) ([]models.Address, error)
	FindByID(ctx context.Context, userID uint, addressID uint) (*models.Address, error)
	Update(ctx context.Context, userID uint, addressID uint, address *models.Address) error
	Delete(ctx context.Context, userID uint, addressID uint) error
	FindAnotherAddress(ctx context.Context, userID uint, excludeID uint) (*models.Address, error)
	SetDefault(ctx context.Context,userID uint,addressID uint) error
}
