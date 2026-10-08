package brand

import (
	"context"
	"ecommerce/internal/delivery/http/validator"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

type BrandUseCase struct {
	brandRepo     interfaces.BrandRepository
	productRepo   interfaces.ProductRepository
	objectStorage interfaces.ObjectStorage
	log           *slog.Logger
}

func NewBrandUseCase(
	brandRepo interfaces.BrandRepository,
	productRepo interfaces.ProductRepository,
	objectStorage interfaces.ObjectStorage,
	log *slog.Logger,
) *BrandUseCase {
	return &BrandUseCase{
		brandRepo:     brandRepo,
		productRepo:   productRepo,
		objectStorage: objectStorage,
		log:           log,
	}
}

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

func stringPtr(value string) *string {
	return &value
}

func (u *BrandUseCase) CreateBrand(ctx context.Context, name string, description string) (*models.Brand, error) {

	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if name == "" {
		return nil, domainerrors.ErrInvalidBrandName
	}

	brand := &models.Brand{
		Name:        name,
		Description: stringPtr(description),
		IsActive:    true,
	}
	if err := u.brandRepo.Create(ctx, brand); err != nil {
		u.log.Error(
			"failed to create brand",
			"name", name,
			"error", err,
		)
		return nil, err
	}
	u.log.Info(
		"brand created",
		"brand_id", brand.ID,
		"name", brand.Name,
	)
	return brand, nil
}
func (u *BrandUseCase) GetBrand(ctx context.Context, id int64) (*models.Brand, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidBrandID
	}
	brand, err := u.brandRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return brand, nil
}
func (u *BrandUseCase) UpdateBrand(ctx context.Context, id int64, name string, description string) (*models.Brand, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidBrandID
	}
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if !validator.ValidateName(name) {
		return nil, domainerrors.ErrInvalidBrandName
	}
	brand, err := u.brandRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existing, err := u.brandRepo.GetByExactName(ctx, name)
	if err == nil && existing != nil && brand.ID != existing.ID {
		return nil, domainerrors.ErrBrandNameAlreadyExists
	}
	brand.Name = name
	brand.Description = stringPtr(description)
	if err := u.brandRepo.Update(ctx, brand); err != nil {
		u.log.Error(
			"failed to update brand",
			"brand_id", id,
			"error", err,
		)
		return nil, err
	}
	u.log.Info(
		"brand updated",
		"brand_id", id,
	)
	return brand, nil
}
func (u *BrandUseCase) DeleteBrand(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainerrors.ErrInvalidBrandID
	}
	_, err := u.brandRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	count, err := u.productRepo.CountActiveByBrandID(
		ctx,
		id,
	)
	if err != nil {
		u.log.Error(
			"failed to check brand products",
			"brand_id", id,
			"error", err,
		)
		return err
	}

	if count > 0 {
		return domainerrors.ErrBrandHasActiveProducts
	}
	if err := u.brandRepo.SoftDelete(ctx, id); err != nil {
		u.log.Error(
			"failed to soft delete brand",
			"brand_id", id,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"brand soft deleted",
		"brand_id", id,
	)
	return nil
}
func (u *BrandUseCase) ListBrands(ctx context.Context, search string, page int, limit int) ([]models.Brand, int64, error) {
	search = strings.TrimSpace(search)
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	brands, total, err := u.brandRepo.List(
		ctx,
		search,
		page,
		limit,
	)
	if err != nil {
		u.log.Error(
			"failed to list brands",
			"search", search,
			"page", page,
			"limit", limit,
			"error", err,
		)
		return nil, 0, err
	}
	return brands, total, nil
}
func (u *BrandUseCase) RestoreBrand(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainerrors.ErrInvalidBrandID
	}
	brand, err := u.brandRepo.GetDeletedByID(ctx, id)
	if err != nil {
		return err
	}
	existing, err := u.brandRepo.GetByExactName(ctx, brand.Name)
	if err == nil && existing != nil {
		return domainerrors.ErrRestoreBrandAlreadyExists
	}
	if err := u.brandRepo.Restore(ctx, id); err != nil {
		u.log.Error(
			"failed to restore brand",
			"brand_id", id,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"brand restored",
		"brand_id", id,
	)
	return nil
}
func (u *BrandUseCase) ListDeletedBrands(ctx context.Context, search string, page int, limit int) ([]models.Brand, int64, error) {
	search = strings.TrimSpace(search)
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	brands, total, err := u.brandRepo.ListDeleted(
		ctx,
		search,
		page,
		limit,
	)
	if err != nil {
		u.log.Error(
			"failed to list deleted brands",
			"search", search,
			"page", page,
			"limit", limit,
			"error", err,
		)
		return nil, 0, err
	}
	return brands, total, nil
}
func (u *BrandUseCase) ToggleBrandActive(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainerrors.ErrInvalidBrandID
	}
	brand, err := u.brandRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.brandRepo.ToggleActive(ctx, id); err != nil {
		u.log.Error(
			"failed to toggle brand active status",
			"brand_id", id,
			"current_status", brand.IsActive,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"brand active status toggled",
		"brand_id", id,
		"previous_status", brand.IsActive,
		"new_status", !brand.IsActive,
	)
	return nil
}
func (u *BrandUseCase) UploadBrandLogo(ctx context.Context, id int64, file io.Reader, size int64, contentType string) (*models.Brand, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidBrandID
	}
	brand, err := u.brandRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return nil, domainerrors.ErrInvalidBrandLogo
	}
	const maxFileSize = 5 * 1024 * 1024 //5mb
	if size <= 0 || size > maxFileSize {
		return nil, domainerrors.ErrInvalidBrandLogoSize
	}
	objectKey := fmt.Sprintf("brands/%d/logo", brand.ID)
	logoURL, err := u.objectStorage.Upload(ctx, objectKey, file, size, contentType)
	if err != nil {
		u.log.Error(
			"failed to upload brand logo",
			"brand_id", id,
			"error", err,
		)
		return nil, err
	}
	if err := u.brandRepo.UpdateLogo(ctx, id, &logoURL); err != nil {
		u.log.Error(
			"failed to update brand logo reference",
			"brand_id", id,
			"error", err,
		)
		return nil, err
	}
	brand.Logo = &logoURL
	u.log.Info(
		"brand logo uploaded",
		"brand_id", id,
	)
	return brand, nil
}
func (u *BrandUseCase) DeleteBrandLogo(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainerrors.ErrInvalidBrandID
	}
	brand, err := u.brandRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if brand.Logo == nil || *brand.Logo == "" {
		return nil
	}
	objectKey := fmt.Sprintf("brands/%d/logo", brand.ID)
	if err := u.objectStorage.Delete(ctx, objectKey); err != nil {
		u.log.Error(
			"failed to delete brand logo",
			"brand_id", id,
			"error", err,
		)
		return err
	}
	if err := u.brandRepo.UpdateLogo(ctx, id, nil); err != nil {
		u.log.Error(
			"failed to clear brand logo",
			"brand_id", id,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"brand logo deleted",
		"brand_id", id,
	)
	return nil
}
