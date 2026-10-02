package interfaces

import (
	"context"

	"ecommerce/internal/domain/models"
)

type AddressUseCase interface {
	AddAddress(ctx context.Context, userID uint, address *models.Address) (*models.Address, error)
	GetAddresses(ctx context.Context, userID uint) ([]models.Address, error)
	UpdateAddress(ctx context.Context, userID uint, addressID uint, address *models.Address) (*models.Address, error)
	DeleteAddress(ctx context.Context, userID uint, addressID uint) error
	SetDefaultAddress(ctx context.Context, userID uint, addressID uint) error
}
