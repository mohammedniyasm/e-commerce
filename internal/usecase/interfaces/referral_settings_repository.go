package interfaces

import (
	"context"
	"ecommerce/internal/domain/models"
)

type ReferralSettingsRepository interface {
	Get(ctx context.Context) (*models.ReferralSettings, error)
	Update(ctx context.Context, settings *models.ReferralSettings) error
}
