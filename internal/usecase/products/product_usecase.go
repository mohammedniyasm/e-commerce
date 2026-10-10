package products

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"unicode"
)

type ProductUseCase struct {
	productRepo  interfaces.ProductRepository
	categoryRepo interfaces.CategoryRepository
	brandRepo    interfaces.BrandRepository
	log          *slog.Logger
}

func NewProductUseCase(
	productRepo interfaces.ProductRepository,
	categoryRepo interfaces.CategoryRepository,
	brandRepo interfaces.BrandRepository,
	log *slog.Logger,
) *ProductUseCase {
	return &ProductUseCase{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
		log:          log,
	}
}

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)
const (
	sortDefault   = ""
	sortPriceAsc  = "price_asc"
	sortPriceDesc = "price_desc"
	sortNameAsc   = "name_asc"
	sortNameDesc  = "name_desc"
)
const (
	defaultStorePage  = 1
	defaultStoreLimit = 12
	maxStoreLimit     = 100
)

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	previousDash := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			previousDash = false
		default:
			if !previousDash {
				builder.WriteRune('-')
				previousDash = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}
func (u *ProductUseCase) validateCategory(ctx context.Context, categoryID int64) error {
	if categoryID <= 0 {
		return domainerrors.ErrInvalidCategoryID
	}
	category, err := u.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}
	if !category.IsActive {
		return domainerrors.ErrCategoryInactive
	}
	return nil
}
func (u *ProductUseCase) validateBrand(ctx context.Context, brandID int64) error {
	if brandID <= 0 {
		return domainerrors.ErrInvalidBrandID
	}
	brand, err := u.brandRepo.GetByID(ctx, brandID)
	if err != nil {
		return err
	}
	if !brand.IsActive {
		return domainerrors.ErrBrandInactive
	}
	return nil
}
func (u *ProductUseCase) generateUniqueSlug(ctx context.Context, name string, productID int64) (string, error) {
	baseSlug := slugify(name)
	if baseSlug == "" {
		return "", domainerrors.ErrInvalidProductName
	}
	for counter := 1; counter <= 1000; counter++ {
		slug := baseSlug
		if counter > 1 {
			slug = baseSlug + "-" + strconv.Itoa(counter)
		}
		existing, err := u.productRepo.GetBySlug(ctx, slug)
		if err != nil {
			if errors.Is(err, domainerrors.ErrRecordNotFound) {
				return slug, nil
			}
			return "", err
		}
		if existing != nil && int64(existing.ID) == productID {
			return slug, nil
		}
	}
	return "", domainerrors.ErrProductSlugExists
}
func (u *ProductUseCase) CreateProduct(ctx context.Context, name string, shortDescription *string, description *string, categoryID int64, brandID int64) (*models.Product, error) {

	name = strings.TrimSpace(name)

	if name == "" {
		return nil, domainerrors.ErrInvalidProductName
	}

	if err := u.validateCategory(ctx, categoryID); err != nil {
		return nil, err
	}

	if err := u.validateBrand(ctx, brandID); err != nil {
		return nil, err
	}

	slug, err := u.generateUniqueSlug(ctx, name, 0)
	if err != nil {
		return nil, err
	}

	product := &models.Product{
		Name:             name,
		Slug:             slug,
		ShortDescription: shortDescription,
		Description:      description,
		CategoryID:       uint(categoryID),
		BrandID:          uint(brandID),
		IsActive:         true,
		IsListed:         true,
	}

	if err := u.productRepo.Create(ctx, product); err != nil {
		u.log.Error(
			"failed to create product",
			"name", name,
			"category_id", categoryID,
			"brand_id", brandID,
			"error", err,
		)
		return nil, err
	}

	u.log.Info(
		"product created",
		"product_id", product.ID,
		"name", product.Name,
	)

	return product, nil
}
func (u *ProductUseCase) GetProduct(ctx context.Context, id uint) (*models.Product, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidProductID
	}
	product, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}
func (u *ProductUseCase) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, domainerrors.ErrInvalidProductSlug
	}
	product, err := u.productRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return product, nil
}
func (u *ProductUseCase) UpdateProduct(ctx context.Context, id uint, name string, shortDescription *string, description *string, categoryID int64, brandID int64) (*models.Product, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidProductID
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domainerrors.ErrInvalidProductName
	}
	if err := u.validateCategory(ctx, categoryID); err != nil {
		return nil, err
	}
	if err := u.validateBrand(ctx, brandID); err != nil {
		return nil, err
	}
	product, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	slug, err := u.generateUniqueSlug(ctx, name, int64(id))
	if err != nil {
		return nil, err
	}

	product.Name = name
	product.Slug = slug
	product.ShortDescription = shortDescription
	product.Description = description
	product.CategoryID = uint(categoryID)
	product.BrandID = uint(brandID)

	if err := u.productRepo.Update(ctx, product); err != nil {
		u.log.Error(
			"failed to update product",
			"product_id", id,
			"error", err,
		)
		return nil, err
	}

	u.log.Info(
		"product updated",
		"product_id", id,
	)

	return product, nil
}
func (u *ProductUseCase) DeleteProduct(ctx context.Context, id uint) error {

	if id <= 0 {
		return domainerrors.ErrInvalidProductID
	}

	_, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.productRepo.SoftDelete(ctx, id); err != nil {
		u.log.Error(
			"failed to soft delete product",
			"product_id", id,
			"error", err,
		)
		return err
	}

	u.log.Info(
		"product soft deleted",
		"product_id", id,
	)

	return nil
}
func (u *ProductUseCase) ListProducts(ctx context.Context, search string, page int, limit int, isActive *bool, isListed *bool, categoryID *int64, brandID *int64) ([]models.Product, int64, error) {

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

	products, total, err := u.productRepo.List(
		ctx,
		search,
		page,
		limit,
		isActive,
		isListed,
		categoryID,
		brandID,
	)
	if err != nil {
		u.log.Error(
			"failed to list products",
			"search", search,
			"page", page,
			"limit", limit,
			"is_active", isActive,
			"is_listed", isListed,
			"error", err,
		)
		return nil, 0, err
	}

	return products, total, nil
}
func (u *ProductUseCase) ListDeletedProducts(ctx context.Context, search string, page int, limit int) ([]models.Product, int64, error) {

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

	products, total, err := u.productRepo.ListDeleted(
		ctx,
		search,
		page,
		limit,
	)
	if err != nil {
		u.log.Error(
			"failed to list deleted products",
			"search", search,
			"page", page,
			"limit", limit,
			"error", err,
		)
		return nil, 0, err
	}

	return products, total, nil
}
func (u *ProductUseCase) RestoreProduct(ctx context.Context, id uint) error {

	if id <= 0 {
		return domainerrors.ErrInvalidProductID
	}

	product, err := u.productRepo.GetDeletedByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.validateCategory(
		ctx,
		int64(product.CategoryID),
	); err != nil {
		return err
	}

	if err := u.validateBrand(
		ctx,
		int64(product.BrandID),
	); err != nil {
		return err
	}

	if err := u.productRepo.Restore(ctx, id); err != nil {
		u.log.Error(
			"failed to restore product",
			"product_id", id,
			"error", err,
		)
		return err
	}

	u.log.Info(
		"product restored",
		"product_id", id,
	)

	return nil
}
func (u *ProductUseCase) ToggleProductActive(ctx context.Context, id uint) error {
	if id <= 0 {
		return domainerrors.ErrInvalidProductID
	}
	product, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.productRepo.ToggleActive(ctx, id); err != nil {
		u.log.Error(
			"failed to toggle product active status",
			"product_id", id,
			"current_status", product.IsActive,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"product active status toggled",
		"product_id", id,
		"previous_status", product.IsActive,
		"new_status", !product.IsActive,
	)
	return nil
}
func (u *ProductUseCase) ToggleProductListed(ctx context.Context, id uint) error {
	if id <= 0 {
		return domainerrors.ErrInvalidProductID
	}
	product, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.productRepo.ToggleListed(ctx, id); err != nil {
		u.log.Error(
			"failed to toggle product listed status",
			"product_id", id,
			"current_status", product.IsListed,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"product listed status toggled",
		"product_id", id,
		"previous_status", product.IsListed,
		"new_status", !product.IsListed,
	)
	return nil
}
func (u *ProductUseCase) ListStoreProducts(
	ctx context.Context,
	search string,
	categoryID *int64,
	brandID *int64,
	minPrice *float64,
	maxPrice *float64,
	sort string,
	page int,
	limit int,
) ([]models.Product, int64, error) {

	search = strings.TrimSpace(search)
	sort = strings.TrimSpace(sort)

	if page < 1 {
		page = defaultStorePage
	}

	if limit < 1 {
		limit = defaultStoreLimit
	}

	if limit > maxStoreLimit {
		limit = maxStoreLimit
	}

	if categoryID != nil && *categoryID <= 0 {
		return nil, 0, domainerrors.ErrInvalidCategoryID
	}

	if brandID != nil && *brandID <= 0 {
		return nil, 0, domainerrors.ErrInvalidBrandID
	}

	if minPrice != nil && *minPrice < 0 {
		return nil, 0, domainerrors.ErrInvalidPriceRange
	}

	if maxPrice != nil && *maxPrice < 0 {
		return nil, 0, domainerrors.ErrInvalidPriceRange
	}

	if minPrice != nil &&
		maxPrice != nil &&
		*minPrice > *maxPrice {

		return nil, 0, domainerrors.ErrInvalidPriceRange
	}

	switch sort {
	case sortDefault,
		sortPriceAsc,
		sortPriceDesc,
		sortNameAsc,
		sortNameDesc:

	default:
		return nil, 0, domainerrors.ErrInvalidProductSort
	}

	products, total, err := u.productRepo.ListStoreProducts(
		ctx,
		search,
		categoryID,
		brandID,
		minPrice,
		maxPrice,
		sort,
		page,
		limit,
	)

	if err != nil {
		u.log.Error(
			"failed to list store products",
			"search", search,
			"category_id", categoryID,
			"brand_id", brandID,
			"min_price", minPrice,
			"max_price", maxPrice,
			"sort", sort,
			"page", page,
			"limit", limit,
			"error", err,
		)

		return nil, 0, err
	}

	u.log.Info(
		"store products listed",
		"search", search,
		"category_id", categoryID,
		"brand_id", brandID,
		"min_price", minPrice,
		"max_price", maxPrice,
		"sort", sort,
		"page", page,
		"limit", limit,
		"total", total,
	)

	return products, total, nil
}
func (u *ProductUseCase) GetStoreProductByID(ctx context.Context, id int64) (*models.Product, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidProductID
	}

	product, err := u.productRepo.GetStoreProductByID(
		ctx,
		id,
	)
	if err != nil {
		u.log.Error(
			"failed to fetch store product",
			"product_id", id,
			"error", err,
		)
		return nil, err
	}

	u.log.Info(
		"store product fetched",
		"product_id", product.ID,
		"slug", product.Slug,
	)

	return product, nil
}
func (u *ProductUseCase) GetStoreProductBySlug(ctx context.Context, slug string) (*models.Product, error) {

	slug = strings.TrimSpace(slug)

	if slug == "" {
		return nil, domainerrors.ErrInvalidProductSlug
	}

	product, err := u.productRepo.GetStoreProductBySlug(
		ctx,
		slug,
	)
	if err != nil {
		u.log.Error(
			"failed to fetch store product by slug",
			"slug", slug,
			"error", err,
		)
		return nil, err
	}

	u.log.Info(
		"store product fetched by slug",
		"product_id", product.ID,
		"slug", product.Slug,
	)

	return product, nil
}
func (u *ProductUseCase) GetRelatedProducts(ctx context.Context,productID int64,limit int) ([]models.Product, error) {

	if productID <= 0 {
		return nil, domainerrors.ErrInvalidProductID
	}

	if limit <= 0 {
		limit = 8
	}

	if limit > 20 {
		limit = 20
	}

	products, err := u.productRepo.GetRelatedProducts(
		ctx,
		productID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get related products for product %d: %w",
			productID,
			err,
		)
	}

	return products, nil
}
