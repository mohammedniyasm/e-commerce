package productimage

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type ProductImageUseCase struct {
	imageRepo     interfaces.ProductImageRepository
	productRepo   interfaces.ProductRepository
	variantRepo   interfaces.ProductVariantRepository
	objectStorage interfaces.ObjectStorage
	log           *slog.Logger
}

func NewProductImageUseCase(
	imageRepo interfaces.ProductImageRepository,
	productRepo interfaces.ProductRepository,
	variantRepo interfaces.ProductVariantRepository,
	objectStorage interfaces.ObjectStorage,
	log *slog.Logger,
) *ProductImageUseCase {
	return &ProductImageUseCase{
		imageRepo:     imageRepo,
		productRepo:   productRepo,
		variantRepo:   variantRepo,
		objectStorage: objectStorage,
		log:           log,
	}
}

const (
	maxImageSize     = 5 * 1024 * 1024
	minProductImages = 3
)

func (u *ProductImageUseCase) validateProduct(ctx context.Context, productID int64) error {

	if productID <= 0 {
		return domainerrors.ErrInvalidProductID
	}

	_, err := u.productRepo.GetByID(ctx, uint(productID))

	return err
}
func (u *ProductImageUseCase) validateVariant(ctx context.Context, productID int64, variantID *int64) error {

	if variantID == nil {
		return nil
	}

	if *variantID <= 0 {
		return domainerrors.ErrInvalidVariantID
	}

	variant, err := u.variantRepo.GetByID(
		ctx,
		*variantID,
	)
	if err != nil {
		return err
	}

	if int64(variant.ProductID) != productID {
		return domainerrors.ErrVariantDoesNotBelongToProduct
	}
	return nil
}
func validateImage(size int64, contentType string) error {

	if size <= 0 || size > maxImageSize {
		return domainerrors.ErrInvalidProductImageSize
	}

	switch contentType {
	case "image/jpeg",
		"image/png",
		"image/webp":
		return nil

	default:
		return domainerrors.ErrInvalidProductImage
	}
}
func (u *ProductImageUseCase) UploadProductImages(ctx context.Context, productID int64, variantID *int64, files []interfaces.ProductImageUpload) ([]models.ProductImage, error) {

	if err := u.validateProduct(ctx, productID); err != nil {
		return nil, err
	}

	if err := u.validateVariant(
		ctx,
		productID,
		variantID,
	); err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, domainerrors.ErrProductImagesRequired
	}
	// productID=
	existingCount, err := u.imageRepo.CountByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, err
	}

	if existingCount+int64(len(files)) < minProductImages {
		return nil, domainerrors.ErrMinimumProductImages
	}

	uploadedImages := make(
		[]models.ProductImage,
		0,
		len(files),
	)

	currentImages, err := u.imageRepo.ListByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, err
	}

	nextDisplayOrder := len(currentImages) + 1

	hasPrimary := false

	for i := range currentImages {
		if currentImages[i].IsPrimary {
			hasPrimary = true
			break
		}
	}

	for i, file := range files {

		if err := validateImage(
			file.Size,
			file.ContentType,
		); err != nil {
			return nil, err
		}

		objectKey := fmt.Sprintf(
			"products/%d/images/%s",
			productID,
			uuid.NewString(),
		)

		imageURL, err := u.objectStorage.Upload(
			ctx,
			objectKey,
			file.File,
			file.Size,
			file.ContentType,
		)
		if err != nil {

			u.log.Error(
				"failed to upload product image",
				"product_id", productID,
				"error", err,
			)

			return nil, err
		}

		isPrimary := false

		if !hasPrimary && i == 0 {
			isPrimary = true
			hasPrimary = true
		}

		image := &models.ProductImage{
			ProductID:    uint(productID),
			VariantID:    nil,
			ImageURL:     imageURL,
			DisplayOrder: nextDisplayOrder,
			IsPrimary:    isPrimary,
		}

		if variantID != nil {
			id := uint(*variantID)
			image.VariantID = &id
		}

		if err := u.imageRepo.Create(ctx, image); err != nil {

			u.log.Error(
				"failed to save product image",
				"product_id", productID,
				"error", err,
			)

			return nil, err
		}

		uploadedImages = append(
			uploadedImages,
			*image,
		)

		nextDisplayOrder++
	}

	u.log.Info(
		"product images uploaded",
		"product_id", productID,
		"image_count", len(uploadedImages),
	)

	return uploadedImages, nil
}
func (u *ProductImageUseCase) GetProductImage(ctx context.Context, id int64) (*models.ProductImage, error) {

	if id <= 0 {
		return nil, domainerrors.ErrInvalidProductImageID
	}

	image, err := u.imageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return image, nil
}
func (u *ProductImageUseCase) ListProductImages(ctx context.Context, productID int64) ([]models.ProductImage, error) {

	if err := u.validateProduct(ctx, productID); err != nil {
		return nil, err
	}

	return u.imageRepo.ListByProductID(
		ctx,
		productID,
	)
}
func (u *ProductImageUseCase) ListVariantImages(ctx context.Context, variantID int64) ([]models.ProductImage, error) {

	if variantID <= 0 {
		return nil, domainerrors.ErrInvalidVariantID
	}

	_, err := u.variantRepo.GetByID(ctx, variantID)
	if err != nil {
		return nil, err
	}

	return u.imageRepo.ListByVariantID(
		ctx,
		variantID,
	)
}
func (u *ProductImageUseCase) UpdateProductImage(ctx context.Context, id int64, variantID *int64, displayOrder *int, isPrimary *bool) (*models.ProductImage, error) {

	if id <= 0 {
		return nil, domainerrors.ErrInvalidProductImageID
	}

	image, err := u.imageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if variantID != nil {
		if *variantID <= 0 {
			return nil, domainerrors.ErrInvalidVariantID
		}
		variant, err := u.variantRepo.GetByID(ctx, *variantID)
		if err != nil {
			return nil, err
		}
		if int64(variant.ProductID) != int64(image.ProductID) {
			return nil, domainerrors.ErrVariantDoesNotBelongToProduct
		}
		newVariantID := uint(*variantID)
		image.VariantID = &newVariantID
	}
	if displayOrder != nil {
		if *displayOrder < 0 {
			return nil, domainerrors.ErrInvalidDisplayOrder
		}
		u.imageRepo.UpdateDisplayOrder(ctx,id,*displayOrder)
		image.DisplayOrder = *displayOrder
	}
	if isPrimary != nil {
		if *isPrimary {

			if err := u.imageRepo.ClearPrimaryByProductID(
				ctx,
				int64(image.ProductID),
				int64(image.ID),
			); err != nil {
				return nil, err
			}
		}
		image.IsPrimary = *isPrimary
	}

	if err := u.imageRepo.Update(
		ctx,
		image,
	); err != nil {
		u.log.Error(
			"failed to update product image",
			"image_id", id,
			"error", err,
		)
		return nil, err
	}

	u.log.Info(
		"product image updated",
		"image_id", id,
		"product_id", image.ProductID,
		"variant_id", image.VariantID,
	)

	return image, nil
}
func (u *ProductImageUseCase) DeleteProductImage(ctx context.Context, id int64) error {

	if id <= 0 {
		return domainerrors.ErrInvalidProductImageID
	}

	image, err := u.imageRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	count, err := u.imageRepo.CountByProductID(
		ctx, int64(image.ProductID),
	)
	if err != nil {
		return err
	}

	if count <= minProductImages {
		return domainerrors.ErrMinimumProductImages
	}

	if err := u.objectStorage.DeleteByURL(
		ctx,
		image.ImageURL,
	); err != nil {
		u.log.Error(
			"failed to delete product image from storage",
			"image_id", id,
			"error", err,
		)
		return err
	}

	if err := u.imageRepo.Delete(ctx, id); err != nil {
		u.log.Error(
			"failed to delete product image record",
			"image_id", id,
			"error", err,
		)
		return err
	}

	u.log.Info(
		"product image deleted",
		"image_id", id,
	)

	return nil
}
