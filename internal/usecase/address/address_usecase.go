package address

import (
	"context"
	"ecommerce/internal/delivery/http/validator"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"log/slog"
	"strings"
)

type AddressUseCase struct {
	addressRepo interfaces.AddressRepository
	log         *slog.Logger
}

func NewAddressUseCase(
	addressRepo interfaces.AddressRepository,
	log *slog.Logger,
) *AddressUseCase {
	return &AddressUseCase{
		addressRepo: addressRepo,
		log:         log,
	}
}
func (u *AddressUseCase) AddAddress(ctx context.Context, userID uint, address *models.Address) (*models.Address, error) {

	if userID == 0 {
		return nil, domainerrors.ErrInvalidUserID
	}

	address.Name = strings.TrimSpace(address.Name)
	address.Phone = strings.TrimSpace(address.Phone)
	address.AddressLine1 = strings.TrimSpace(address.AddressLine1)
	address.AddressLine2 = strings.TrimSpace(address.AddressLine2)
	address.City = strings.TrimSpace(address.City)
	address.PostalCode = strings.TrimSpace(address.PostalCode)
	address.State = strings.TrimSpace(address.State)
	address.Country = strings.TrimSpace(address.Country)

	if !validator.ValidateName(address.Name) {
		return nil, domainerrors.ErrInvalidName
	}

	if !validator.ValidatePhone(address.Phone) {
		return nil, domainerrors.ErrInvalidPhone
	}

	if address.AddressLine1 == "" ||
		address.City == "" ||
		address.PostalCode == "" ||
		address.State == "" {
		return nil, domainerrors.ErrInvalidAddress
	}

	if address.Country == "" {
		address.Country = "India"
	}

	address.UserID = userID

	count, err := u.addressRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		address.IsDefault = true
	}
	if address.IsDefault && count > 0 {
		if err := u.addressRepo.ClearDefaultByUserID(ctx, userID); err != nil {
			return nil, err
		}
	}

	if err := u.addressRepo.Create(ctx, address); err != nil {
		u.log.Error(
			"failed to create address",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}

	u.log.Info(
		"address created",
		"user_id", userID,
		"address_id", address.ID,
	)

	return address, nil
}
func (u *AddressUseCase) GetAddresses(ctx context.Context, userID uint) ([]models.Address, error) {
	if userID == 0 {
		return nil, domainerrors.ErrInvalidUserID
	}

	return u.addressRepo.GetByUserID(ctx, userID)
}
func (u *AddressUseCase) UpdateAddress(ctx context.Context, userID uint, addressID uint, address *models.Address) (*models.Address, error) {

	if userID == 0 || addressID == 0 {
		return nil, domainerrors.ErrInvalidUserID
	}

	address.Name = strings.TrimSpace(address.Name)
	address.Phone = strings.TrimSpace(address.Phone)
	address.AddressLine1 = strings.TrimSpace(address.AddressLine1)
	address.AddressLine2 = strings.TrimSpace(address.AddressLine2)
	address.City = strings.TrimSpace(address.City)
	address.PostalCode = strings.TrimSpace(address.PostalCode)
	address.State = strings.TrimSpace(address.State)
	address.Country = strings.TrimSpace(address.Country)

	if !validator.ValidateName(address.Name) {
		return nil, domainerrors.ErrInvalidName
	}

	if !validator.ValidatePhone(address.Phone) {
		return nil, domainerrors.ErrInvalidPhone
	}

	if address.AddressLine1 == "" ||
		address.City == "" ||
		address.PostalCode == "" ||
		address.State == "" {
		return nil, domainerrors.ErrInvalidAddress
	}

	if address.Country == "" {
		address.Country = "India"
	}

	existingAddress, err := u.addressRepo.FindByID(
		ctx,
		userID,
		addressID,
	)
	if err != nil {
		return nil, err
	}

	if address.IsDefault && !existingAddress.IsDefault {
		if err := u.addressRepo.ClearDefaultByUserID(
			ctx,
			userID,
		); err != nil {
			return nil, err
		}
	}

	address.UserID = userID
	address.ID = addressID

	if err := u.addressRepo.Update(
		ctx,
		userID,
		addressID,
		address,
	); err != nil {
		return nil, err
	}

	u.log.Info(
		"address updated",
		"user_id", userID,
		"address_id", addressID,
	)

	return address, nil
}
func (u *AddressUseCase) DeleteAddress(ctx context.Context, userID uint, addressID uint) error {

	if userID == 0 || addressID == 0 {
		return domainerrors.ErrInvalidUserID
	}

	address, err := u.addressRepo.FindByID(
		ctx,
		userID,
		addressID,
	)
	if err != nil {
		return err
	}
	if err := u.addressRepo.Delete(ctx, userID, addressID); err != nil {
		u.log.Error(
			"failed to delete address",
			"user_id", userID,
			"address_id", addressID,
			"error", err,
		)
		return err
	}
	if address.IsDefault {
		anotherAddress, err := u.addressRepo.FindAnotherAddress(
			ctx,
			userID,
			addressID,
		)

		if err == nil {
			if err := u.addressRepo.ClearDefaultByUserID(ctx, userID); err != nil {
				return err
			}
			anotherAddress.IsDefault = true
			if err := u.addressRepo.Update(ctx, userID, anotherAddress.ID, anotherAddress); err != nil {
				return err
			}
		}
	}

	u.log.Info(
		"address deleted",
		"user_id", userID,
		"address_id", addressID,
	)

	return nil
}
func (u *AddressUseCase) SetDefaultAddress(ctx context.Context,userID uint,addressID uint) error {
	if userID == 0 || addressID == 0 {
		return domainerrors.ErrInvalidUserID
	}

	if err := u.addressRepo.SetDefault(
		ctx,
		userID,
		addressID,
	); err != nil {
		return err
	}

	u.log.Info(
		"default address updated",
		"user_id", userID,
		"address_id", addressID,
	)

	return nil
}
