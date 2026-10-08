package category

import (
	"context"
	"ecommerce/internal/delivery/http/validator"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"log/slog"
	"strings"
)

type CategoryUseCase struct {
	categoryRepo interfaces.CategoryRepository
	productRepo  interfaces.ProductRepository
	log          *slog.Logger
}

func NewCategoryUseCase(
	categoryRepo interfaces.CategoryRepository,
	productRepo interfaces.ProductRepository,
	log *slog.Logger,
) *CategoryUseCase {
	return &CategoryUseCase{
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
		log:          log,
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
func (u *CategoryUseCase) CreateCategory(ctx context.Context, name string, description string) (*models.Category, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" {
		return nil, domainerrors.ErrInvalidCategoryName
	}
	category := &models.Category{
		Name:        name,
		Description: stringPtr(description),
		IsActive:    true,
	}
	if err := u.categoryRepo.Create(ctx, category); err != nil {
		u.log.Error(
			"failed to create category",
			"name", name,
			"error", err,
		)
		return nil, err
	}
	u.log.Info(
		"category created",
		"category_id", category.ID,
		"name", category.Name,
	)
	return category, nil
}
func (u *CategoryUseCase) GetCategory(ctx context.Context, id int64) (*models.Category, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidCategoryID
	}
	category, err := u.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return category, nil
}
func (u *CategoryUseCase) UpdateCategory(ctx context.Context, id int64, name string, description string) (*models.Category, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidCategoryID
	}
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if !validator.ValidateName(name) {
		return nil, domainerrors.ErrInvalidCategoryName
	}

	category, err := u.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existing, err := u.categoryRepo.GetByExactName(ctx, name)
	if err == nil && existing != nil && category.ID != existing.ID {
		return nil, domainerrors.ErrCategoryNameAlreadyExists
	}
	category.Name = name
	category.Description = stringPtr(description)
	if err := u.categoryRepo.Update(ctx, category); err != nil {
		u.log.Error(
			"failed to update category",
			"category_id", id,
			"error", err,
		)
		return nil, err
	}
	u.log.Info(
		"category updated",
		"category_id", id,
	)
	return category, nil
}
func (u *CategoryUseCase) DeleteCategory(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainerrors.ErrInvalidCategoryID
	}
	_, err := u.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	count, err := u.productRepo.CountActiveByCategoryID(
		ctx,
		id,
	)
	if err != nil {
		u.log.Error(
			"failed to check category products",
			"category_id", id,
			"error", err,
		)
		return err
	}

	if count > 0 {
		return domainerrors.ErrCategoryHasActiveProducts
	}
	if err := u.categoryRepo.SoftDelete(ctx, id); err != nil {
		u.log.Error(
			"failed to soft delete category",
			"category_id", id,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"category soft deleted",
		"category_id", id,
	)
	return nil
}
func (u *CategoryUseCase) ListCategories(ctx context.Context, search string, page int, limit int) ([]models.Category, int64, error) {
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
	categories, total, err := u.categoryRepo.List(ctx, search, page, limit)
	if err != nil {
		u.log.Error(
			"failed to list categories",
			"search", search,
			"page", page,
			"limit", limit,
			"error", err,
		)
		return nil, 0, err
	}
	return categories, total, nil
}
func (u *CategoryUseCase) RestoreCategory(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainerrors.ErrInvalidCategoryID
	}
	category, err := u.categoryRepo.GetDeletedByID(ctx, id)
	if err != nil {
		return err
	}
	existing, err := u.categoryRepo.GetByExactName(ctx, category.Name)
	if err == nil && existing != nil {
		return domainerrors.ErrRestoreCategoryAlreadyExists
	}
	if err := u.categoryRepo.Restore(ctx, id); err != nil {
		u.log.Error(
			"failed to restore category",
			"category_id", id,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"category restored",
		"category_id", id,
	)
	return nil
}
func (u *CategoryUseCase) ListDeletedCategories(ctx context.Context, search string, page int, limit int) ([]models.Category, int64, error) {
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
	categories, total, err := u.categoryRepo.ListDeleted(
		ctx,
		search,
		page,
		limit,
	)
	if err != nil {
		u.log.Error(
			"failed to list deleted categories",
			"search", search,
			"page", page,
			"limit", limit,
			"error", err,
		)
		return nil, 0, err
	}

	return categories, total, nil
}
func (u *CategoryUseCase) ToggleCategoryActive(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainerrors.ErrInvalidCategoryID
	}
	category, err := u.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.categoryRepo.ToggleActive(ctx, id); err != nil {
		u.log.Error(
			"failed to toggle category active status",
			"category_id", id,
			"current_status", category.IsActive,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"category active status toggled",
		"category_id", id,
		"previous_status", category.IsActive,
		"new_status", !category.IsActive,
	)
	return nil
}
