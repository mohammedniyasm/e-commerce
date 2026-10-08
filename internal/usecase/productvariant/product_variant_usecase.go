package productvariant

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
)

type ProductVariantUseCase struct {
	variantRepo interfaces.ProductVariantRepository
	productRepo interfaces.ProductRepository
	log         *slog.Logger
}

func NewProductVariantUseCase(
	variantRepo interfaces.ProductVariantRepository,
	productRepo interfaces.ProductRepository,
	log *slog.Logger,
) *ProductVariantUseCase {
	return &ProductVariantUseCase{
		variantRepo: variantRepo,
		productRepo: productRepo,
		log:         log,
	}
}

func normalizeSKU(sku string) string {
	return strings.ToUpper(strings.TrimSpace(sku))
}
func normalizeOptionalValue(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)

	if trimmed == "" {
		return nil
	}

	return &trimmed
}
func (u *ProductVariantUseCase) validateProduct(ctx context.Context, productID int64) error {
	if productID <= 0 {
		return domainerrors.ErrInvalidProductID
	}
	_, err := u.productRepo.GetByID(ctx, uint(productID))
	return err
}
func (u *ProductVariantUseCase) CreateVariant(ctx context.Context, productID int64, size *string, color *string, sku string, mrp float64, sellingPrice float64, stock int) (*models.ProductVariant, error) {

	if err := u.validateProduct(ctx, productID); err != nil {
		return nil, err
	}

	size = normalizeOptionalValue(size)
	color = normalizeOptionalValue(color)
	sku = normalizeSKU(sku)

	if sku == "" {
		return nil, domainerrors.ErrInvalidVariantSKU
	}

	if mrp <= 0 {
		return nil, domainerrors.ErrInvalidMRP
	}

	if sellingPrice <= 0 {
		return nil, domainerrors.ErrInvalidSellingPrice
	}

	if sellingPrice > mrp {
		return nil, domainerrors.ErrSellingPriceGreaterThanMRP
	}

	if stock < 0 {
		return nil, domainerrors.ErrInvalidStock
	}

	existingSKU, err := u.variantRepo.GetBySKU(ctx, sku)
	if err == nil && existingSKU != nil {
		return nil, domainerrors.ErrVariantSKUAlreadyExists
	}

	if err != nil && !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, err
	}

	existingVariant, err := u.variantRepo.GetByProductAndAttributes(
		ctx,
		productID,
		size,
		color,
	)

	if err == nil && existingVariant != nil {
		return nil, domainerrors.ErrVariantAlreadyExists
	}

	if err != nil && !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, err
	}

	variant := &models.ProductVariant{
		ProductID:    uint(productID),
		Size:         size,
		Color:        color,
		SKU:          sku,
		MRP:          mrp,
		SellingPrice: sellingPrice,
		Stock:        stock,
		IsActive:     true,
	}

	if err := u.variantRepo.Create(ctx, variant); err != nil {
		u.log.Error(
			"failed to create product variant",
			"product_id", productID,
			"sku", sku,
			"error", err,
		)
		return nil, err
	}

	u.log.Info(
		"product variant created",
		"variant_id", variant.ID,
		"product_id", productID,
		"sku", variant.SKU,
	)

	return variant, nil
}
func (u *ProductVariantUseCase) GetVariant(ctx context.Context, id int64) (*models.ProductVariant, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidVariantID
	}
	variant, err := u.variantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return variant, nil
}
func (u *ProductVariantUseCase) UpdateVariant(ctx context.Context, id int64, size *string, color *string, sku string, mrp float64, sellingPrice float64, stock int, isActive *bool) (*models.ProductVariant, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidVariantID
	}

	variant, err := u.variantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	size = normalizeOptionalValue(size)
	color = normalizeOptionalValue(color)
	sku = normalizeSKU(sku)

	if sku == "" {
		return nil, domainerrors.ErrInvalidVariantSKU
	}

	if mrp <= 0 {
		return nil, domainerrors.ErrInvalidMRP
	}

	if sellingPrice <= 0 {
		return nil, domainerrors.ErrInvalidSellingPrice
	}

	if sellingPrice > mrp {
		return nil, domainerrors.ErrSellingPriceGreaterThanMRP
	}

	if stock < 0 {
		return nil, domainerrors.ErrInvalidStock
	}
	existingSKU, err := u.variantRepo.GetBySKU(ctx, sku)
	if err == nil && existingSKU != nil && existingSKU.ID != variant.ID {
		return nil, domainerrors.ErrVariantSKUAlreadyExists
	}
	if err != nil && !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, err
	}
	existingVariant, err := u.variantRepo.GetByProductAndAttributes(
		ctx,
		int64(variant.ProductID),
		size,
		color,
	)
	if err == nil &&
		existingVariant != nil &&
		existingVariant.ID != variant.ID {
		return nil, domainerrors.ErrVariantAlreadyExists
	}

	if err != nil && !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, err
	}
	if size != nil {
		variant.Size = size
	}
	if color != nil {
		variant.Color = color
	}
	if sku != "" {
		variant.SKU = sku
	}
	variant.MRP = mrp
	variant.SellingPrice = sellingPrice
	// if stock != nil{
	// }
	variant.Stock = stock
	if isActive != nil{
		variant.IsActive = *isActive
	}
	if err := u.variantRepo.Update(ctx, variant); err != nil {
		u.log.Error(
			"failed to update product variant",
			"variant_id", id,
			"error", err,
		)
		return nil, err
	}

	u.log.Info(
		"product variant updated",
		"variant_id", id,
		"sku", variant.SKU,
	)

	return variant, nil
}
func (u *ProductVariantUseCase) DeleteVariant(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainerrors.ErrInvalidVariantID
	}
	_, err := u.variantRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.variantRepo.Delete(ctx, id); err != nil {
		u.log.Error(
			"failed to delete product variant",
			"variant_id", id,
			"error", err,
		)
		return err
	}
	u.log.Info(
		"product variant deleted",
		"variant_id", id,
	)
	return nil
}
func (u *ProductVariantUseCase) ListVariantsByProductID(ctx context.Context, productID int64) ([]models.ProductVariant, error) {
	if err := u.validateProduct(ctx, productID); err != nil {
		return nil, err
	}
	variants, err := u.variantRepo.ListByProductID(
		ctx,
		productID,
	)
	if err != nil {
		u.log.Error(
			"failed to list product variants",
			"product_id", productID,
			"error", err,
		)
		return nil, err
	}
	return variants, nil
}
