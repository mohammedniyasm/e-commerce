package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type AuthUseCase interface {
	Register(ctc context.Context, user *models.User) (*models.User, error)
}
