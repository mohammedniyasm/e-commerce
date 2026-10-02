package postgres

import (
	"context"

	"ecommerce/internal/domain/models"

	"gorm.io/gorm"
)

type AddressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{
		db: db,
	}
}
func (r *AddressRepository) Create(ctx context.Context, address *models.Address) error {
	return r.db.WithContext(ctx).Create(address).Error
}
func (r *AddressRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&models.Address{}).Where("user_id = ?", userID).Count(&count).Error

	return count, err
}
func (r *AddressRepository) ClearDefaultByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&models.Address{}).Where("user_id = ? AND is_default = ?", userID, true).Update("is_default", false).Error
}
func (r *AddressRepository) GetByUserID(ctx context.Context, userID uint) ([]models.Address, error) {
	var addresses []models.Address
	err := r.db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addresses).
		Error
	return addresses, err
}
func (r *AddressRepository) FindByID(ctx context.Context, userID uint, addressID uint) (*models.Address, error) {
	var address models.Address

	err := r.db.
		WithContext(ctx).
		Where("id = ? AND user_id = ?", addressID, userID).
		First(&address).
		Error

	if err != nil {
		return nil, err
	}

	return &address, nil
}
func (r *AddressRepository) Update(ctx context.Context, userID uint, addressID uint, address *models.Address) error {
	result := r.db.
		WithContext(ctx).
		Model(&models.Address{}).
		Where("id = ? AND user_id = ?", addressID, userID).
		Updates(map[string]interface{}{
			"name":           address.Name,
			"phone":          address.Phone,
			"address_line1":  address.AddressLine1,
			"address_line_2": address.AddressLine2,
			"city":           address.City,
			"postal_code":    address.PostalCode,
			"state":          address.State,
			"country":        address.Country,
			"is_default":     address.IsDefault,
		})
	return result.Error
}
func (r *AddressRepository) FindAnotherAddress(ctx context.Context, userID uint, excludeID uint) (*models.Address, error) {
	var address models.Address

	err := r.db.
		WithContext(ctx).
		Where("user_id = ? AND id <> ?", userID, excludeID).
		Order("created_at DESC").
		First(&address).
		Error

	if err != nil {
		return nil, err
	}

	return &address, nil
}
func (r *AddressRepository) Delete(ctx context.Context, userID uint, addressID uint) error {
	result := r.db.
		WithContext(ctx).
		Where("id = ? AND user_id = ?", addressID, userID).
		Delete(&models.Address{})
	return result.Error
}
func (r *AddressRepository) SetDefault(ctx context.Context,userID uint,addressID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var address models.Address
		if err := tx.
			Where("id = ? AND user_id = ?", addressID, userID).
			First(&address).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.Address{}).
			Where("user_id = ?", userID).
			Update("is_default", false).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.Address{}).
			Where("id = ? AND user_id = ?", addressID, userID).
			Update("is_default", true).
			Error; err != nil {
			return err
		}
		return nil
	})
}
