package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"

	"gorm.io/gorm"
)

type referralSettingsRepository struct {
	db *gorm.DB
}

func NewReferralSettingsRepository(db *gorm.DB) *referralSettingsRepository {
	return &referralSettingsRepository{db: db}
}

func (r *referralSettingsRepository) Get(ctx context.Context) (*models.ReferralSettings, error) {
	var settings models.ReferralSettings
	err := r.db.WithContext(ctx).First(&settings, 1).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &settings, nil
}
func (r *referralSettingsRepository) Update(ctx context.Context, settings *models.ReferralSettings) error {
	return r.db.WithContext(ctx).
		Model(&models.ReferralSettings{}).
		Where("id = ?", settings.ID).
		Select(
			"ReferrerReward",
			"ReferredReward",
			"MinimumOrderAmount",
			"IsActive",
		).
		Updates(settings).Error
}

var _ interfaces.ReferralSettingsRepository = (*referralSettingsRepository)(nil)
