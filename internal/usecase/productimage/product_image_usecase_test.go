package productimage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidateImage(t *testing.T) {
	tests := []struct {
		name        string
		size        int64
		contentType string
		wantError   error
	}{
		// Valid cases
		{
			name:        "valid JPEG image",
			size:        1024,
			contentType: "image/jpeg",
			wantError:   nil,
		},
		{
			name:        "valid PNG image",
			size:        2048,
			contentType: "image/png",
			wantError:   nil,
		},
		{
			name:        "valid WEBP image",
			size:        4096,
			contentType: "image/webp",
			wantError:   nil,
		},
		{
			name:        "image size exactly at maximum limit",
			size:        maxImageSize,
			contentType: "image/jpeg",
			wantError:   nil,
		},

		// Invalid size cases
		{
			name:        "zero image size",
			size:        0,
			contentType: "image/jpeg",
			wantError:   domainerrors.ErrInvalidProductImageSize,
		},
		{
			name:        "negative image size",
			size:        -1,
			contentType: "image/jpeg",
			wantError:   domainerrors.ErrInvalidProductImageSize,
		},
		{
			name:        "image size exceeds maximum limit",
			size:        maxImageSize + 1,
			contentType: "image/jpeg",
			wantError:   domainerrors.ErrInvalidProductImageSize,
		},

		// Invalid content type cases
		{
			name:        "GIF image not allowed",
			size:        1024,
			contentType: "image/gif",
			wantError:   domainerrors.ErrInvalidProductImage,
		},
		{
			name:        "PDF file not allowed",
			size:        1024,
			contentType: "application/pdf",
			wantError:   domainerrors.ErrInvalidProductImage,
		},
		{
			name:        "empty content type",
			size:        1024,
			contentType: "",
			wantError:   domainerrors.ErrInvalidProductImage,
		},

		// Multiple invalid conditions
		{
			name:        "invalid size and content type",
			size:        0,
			contentType: "application/pdf",
			wantError:   domainerrors.ErrInvalidProductImageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImage(tt.size, tt.contentType)

			if !errors.Is(err, tt.wantError) {
				t.Errorf(
					"validateImage() error = %v; want %v",
					err,
					tt.wantError,
				)
			}
		})
	}
}

type productImageFixture struct {
	uc          *ProductImageUseCase
	imageRepo   *mocks.MockProductImageRepository
	productRepo *mocks.MockProductRepository
	variantRepo *mocks.MockProductVariantRepository
	storage     *mocks.MockObjectStorage
}

func newProductImageFixture(t *testing.T) *productImageFixture {
	t.Helper()
	imageRepo := mocks.NewMockProductImageRepository(t)
	productRepo := mocks.NewMockProductRepository(t)
	variantRepo := mocks.NewMockProductVariantRepository(t)
	storage := mocks.NewMockObjectStorage(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &productImageFixture{
		uc:          NewProductImageUseCase(imageRepo, productRepo, variantRepo, storage, log),
		imageRepo:   imageRepo,
		productRepo: productRepo,
		variantRepo: variantRepo,
		storage:     storage,
	}
}

func TestProductImageUseCase_UploadProductImages(t *testing.T) {
	ctx := context.Background()
	validUpload := interfaces.ProductImageUpload{
		File:        bytes.NewReader([]byte("img-bytes")),
		Size:        1024,
		ContentType: "image/jpeg",
	}

	t.Run("validation errors", func(t *testing.T) {
		f := newProductImageFixture(t)
		f.productRepo.EXPECT().GetByID(ctx, uint(1)).Return(&models.Product{ID: 1}, nil).Maybe()

		_, err := f.uc.UploadProductImages(ctx, 0, nil, []interfaces.ProductImageUpload{validUpload})
		assert.ErrorIs(t, err, domainerrors.ErrInvalidProductID)

		_, err = f.uc.UploadProductImages(ctx, 1, nil, nil)
		assert.ErrorIs(t, err, domainerrors.ErrProductImagesRequired)

		f.imageRepo.EXPECT().CountByProductID(ctx, int64(1)).Return(int64(2), nil).Maybe()
		f.imageRepo.EXPECT().ListByProductID(ctx, int64(1)).Return([]models.ProductImage{}, nil).Maybe()
		badUpload := interfaces.ProductImageUpload{File: bytes.NewReader([]byte("x")), Size: 100, ContentType: "text/plain"}
		_, err = f.uc.UploadProductImages(ctx, 1, nil, []interfaces.ProductImageUpload{badUpload})
		assert.ErrorIs(t, err, domainerrors.ErrInvalidProductImage)
	})

	t.Run("variant ownership check", func(t *testing.T) {
		f := newProductImageFixture(t)
		varID := int64(55)
		f.productRepo.EXPECT().GetByID(ctx, uint(1)).Return(&models.Product{ID: 1}, nil).Once()
		f.variantRepo.EXPECT().GetByID(ctx, int64(55)).Return(&models.ProductVariant{ID: 55, ProductID: 999}, nil).Once()

		_, err := f.uc.UploadProductImages(ctx, 1, &varID, []interfaces.ProductImageUpload{validUpload})
		assert.ErrorIs(t, err, domainerrors.ErrVariantDoesNotBelongToProduct)
	})

	t.Run("minimum images check and success", func(t *testing.T) {
		f := newProductImageFixture(t)
		f.productRepo.EXPECT().GetByID(ctx, uint(1)).Return(&models.Product{ID: 1}, nil).Once()
		f.imageRepo.EXPECT().CountByProductID(ctx, int64(1)).Return(int64(0), nil).Once()

		_, err := f.uc.UploadProductImages(ctx, 1, nil, []interfaces.ProductImageUpload{validUpload})
		assert.ErrorIs(t, err, domainerrors.ErrMinimumProductImages)

		f2 := newProductImageFixture(t)
		f2.productRepo.EXPECT().GetByID(ctx, uint(1)).Return(&models.Product{ID: 1}, nil).Once()
		f2.imageRepo.EXPECT().CountByProductID(ctx, int64(1)).Return(int64(2), nil).Once()
		f2.imageRepo.EXPECT().ListByProductID(ctx, int64(1)).Return([]models.ProductImage{}, nil).Maybe()
		f2.storage.EXPECT().Upload(ctx, mock.AnythingOfType("string"), mock.Anything, int64(1024), "image/jpeg").
			Return("https://cdn.example.com/1.jpg", nil).Once()
		f2.imageRepo.EXPECT().Create(ctx, mock.MatchedBy(func(img *models.ProductImage) bool {
			return img.ProductID == 1 && img.ImageURL == "https://cdn.example.com/1.jpg"
		})).Return(nil).Once()

		got, err := f2.uc.UploadProductImages(ctx, 1, nil, []interfaces.ProductImageUpload{validUpload})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "https://cdn.example.com/1.jpg", got[0].ImageURL)
	})
}

func TestProductImageUseCase_GetListUpdateDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("GetProductImage ListProductImages ListVariantImages", func(t *testing.T) {
		f := newProductImageFixture(t)
		_, err := f.uc.GetProductImage(ctx, 0)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidProductImageID)

		img := &models.ProductImage{ID: 3, ProductID: 1}
		f.imageRepo.EXPECT().GetByID(ctx, int64(3)).Return(img, nil).Once()
		got, err := f.uc.GetProductImage(ctx, 3)
		require.NoError(t, err)
		assert.Equal(t, img, got)

		_, err = f.uc.ListProductImages(ctx, 0)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidProductID)
		f.productRepo.EXPECT().GetByID(ctx, uint(1)).Return(&models.Product{ID: 1}, nil).Once()
		f.imageRepo.EXPECT().ListByProductID(ctx, int64(1)).Return([]models.ProductImage{*img}, nil).Once()
		gotList, err := f.uc.ListProductImages(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, gotList, 1)

		_, err = f.uc.ListVariantImages(ctx, 0)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidVariantID)
		f.variantRepo.EXPECT().GetByID(ctx, int64(2)).Return(&models.ProductVariant{ID: 2}, nil).Once()
		f.imageRepo.EXPECT().ListByVariantID(ctx, int64(2)).Return([]models.ProductImage{*img}, nil).Once()
		gotVarList, err := f.uc.ListVariantImages(ctx, 2)
		require.NoError(t, err)
		assert.Len(t, gotVarList, 1)
	})

	t.Run("UpdateProductImage", func(t *testing.T) {
		f := newProductImageFixture(t)
		negOrder := -1
		_, err := f.uc.UpdateProductImage(ctx, 0, nil, nil, nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidProductImageID)

		existing := &models.ProductImage{ID: 3, ProductID: 1, DisplayOrder: 0}
		f.imageRepo.EXPECT().GetByID(ctx, int64(3)).Return(existing, nil).Once()
		_, err = f.uc.UpdateProductImage(ctx, 3, nil, &negOrder, nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidDisplayOrder)

		order := 2
		primary := true
		f.imageRepo.EXPECT().GetByID(ctx, int64(3)).Return(existing, nil).Once()
		f.imageRepo.EXPECT().ClearPrimaryByProductID(ctx, int64(1), int64(3)).Return(nil).Once()
		f.imageRepo.EXPECT().UpdateDisplayOrder(ctx, int64(3), 2).Return(nil).Once()
		f.imageRepo.EXPECT().Update(ctx, mock.MatchedBy(func(img *models.ProductImage) bool {
			return img.ID == 3 && img.IsPrimary
		})).Return(nil).Once()

		got, err := f.uc.UpdateProductImage(ctx, 3, nil, &order, &primary)
		require.NoError(t, err)
		assert.True(t, got.IsPrimary)
	})

	t.Run("DeleteProductImage", func(t *testing.T) {
		f := newProductImageFixture(t)
		assert.ErrorIs(t, f.uc.DeleteProductImage(ctx, 0), domainerrors.ErrInvalidProductImageID)

		img := &models.ProductImage{ID: 3, ProductID: 1, ImageURL: "https://cdn.example.com/3.jpg"}
		f.imageRepo.EXPECT().GetByID(ctx, int64(3)).Return(img, nil).Once()
		f.imageRepo.EXPECT().CountByProductID(ctx, int64(1)).Return(int64(3), nil).Once()
		assert.ErrorIs(t, f.uc.DeleteProductImage(ctx, 3), domainerrors.ErrMinimumProductImages)

		f.imageRepo.EXPECT().GetByID(ctx, int64(3)).Return(img, nil).Once()
		f.imageRepo.EXPECT().CountByProductID(ctx, int64(1)).Return(int64(4), nil).Once()
		f.imageRepo.EXPECT().Delete(ctx, int64(3)).Return(nil).Once()
		f.storage.EXPECT().DeleteByURL(ctx, "https://cdn.example.com/3.jpg").Return(nil).Once()

		require.NoError(t, f.uc.DeleteProductImage(ctx, 3))
	})
}
