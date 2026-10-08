package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

type ProductImageHandler struct {
	imageUseCase interfaces.ProductImageUseCase
}

func NewProductImageHandler(
	imageUseCase interfaces.ProductImageUseCase,
) *ProductImageHandler {
	return &ProductImageHandler{
		imageUseCase: imageUseCase,
	}
}
func (h *ProductImageHandler) handleProductImageError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, domainerrors.ErrInvalidProductID):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})

	case errors.Is(err, domainerrors.ErrInvalidVariantID):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid variant id",
		})

	case errors.Is(err, domainerrors.ErrInvalidProductImageID):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid image id",
		})

	case errors.Is(err, domainerrors.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "product, variant, or image not found",
		})

	case errors.Is(
		err,
		domainerrors.ErrVariantDoesNotBelongToProduct,
	):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "variant does not belong to this product",
		})

	case errors.Is(err, domainerrors.ErrInvalidProductImage):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product image",
		})

	case errors.Is(err, domainerrors.ErrInvalidProductImageSize):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "product image size is invalid",
		})

	case errors.Is(err, domainerrors.ErrProductImagesRequired):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "product images are required",
		})

	case errors.Is(err, domainerrors.ErrMinimumProductImages):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "product must have at least 3 images",
		})

	case errors.Is(err, domainerrors.ErrInvalidDisplayOrder):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid display order",
		})

	default:
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
	}
}
func (h *ProductImageHandler) UploadProductImages(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}
	var variantID *int64
	variantIDValue := strings.TrimSpace(
		c.PostForm("variant_id"),
	)

	if variantIDValue != "" {

		parsedID, err := strconv.ParseInt(
			variantIDValue,
			10,
			64,
		)

		if err != nil || parsedID <= 0 {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid variant id",
			})
			return
		}
		variantID = &parsedID
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid multipart form",
			Error:   err.Error(),
		})
		return
	}
	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "product images are required",
		})
		return
	}
	uploads := make([]interfaces.ProductImageUpload, 0, len(files))
	openedFiles := make([]interface{ Close() error }, 0, len(files))
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			for _, openedFile := range openedFiles {
				_ = openedFile.Close()
			}
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to open product image",
			})
			return
		}
		openedFiles = append(openedFiles, file)
		uploads = append(uploads, interfaces.ProductImageUpload{
			File:        file,
			Size:        fileHeader.Size,
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
		defer func() {
			for _, openedFile := range openedFiles {
				_ = openedFile.Close()
			}
		}()
	}
	images, err := h.imageUseCase.UploadProductImages(c.Request.Context(), int64(productID), variantID, uploads)
	if err != nil {
		h.handleProductImageError(c, err, "failed to upload product images")
		return
	}
	imageResponses := make([]response.ProductImageResponse, 0, len(images))
	for i := range images {
		imageResponses = append(imageResponses, response.ToProductImageResponse(&images[i]))
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "product images uploaded successfully",
		Data:    imageResponses,
	})

}
func (h *ProductImageHandler) ListProductImages(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}
	images, err := h.imageUseCase.ListProductImages(c.Request.Context(), int64(productID))
	if err != nil {
		h.handleProductImageError(c, err, "failed to fetch product images")
		return
	}
	imageResponses := make([]response.ProductImageResponse, 0, len(images))
	for i := range images {
		imageResponses = append(imageResponses, response.ToProductImageResponse(&images[i]))
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product images fetched successfully",
		Data:    imageResponses,
	})

}
func (h *ProductImageHandler) ListVariantImages(c *gin.Context) {

	variantID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil || variantID <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid variant id",
		})
		return
	}

	images, err := h.imageUseCase.ListVariantImages(
		c.Request.Context(),
		variantID,
	)

	if err != nil {
		h.handleProductImageError(
			c,
			err,
			"failed to fetch variant images",
		)
		return
	}

	imageResponses := make(
		[]response.ProductImageResponse,
		0,
		len(images),
	)

	for i := range images {
		imageResponses = append(
			imageResponses,
			response.ToProductImageResponse(&images[i]),
		)
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "variant images fetched successfully",
		Data:    imageResponses,
	})
}
func (h *ProductImageHandler) GetProductImage(c *gin.Context) {

	id, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid image id",
		})
		return
	}

	image, err := h.imageUseCase.GetProductImage(
		c.Request.Context(),
		id,
	)

	if err != nil {
		h.handleProductImageError(
			c,
			err,
			"failed to fetch product image",
		)
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product image fetched successfully",
		Data:    response.ToProductImageResponse(image),
	})
}
func (h *ProductImageHandler) UpdateProductImage(c *gin.Context) {

	id, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid image id",
		})
		return
	}

	var req request.UpdateProductImageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	image, err := h.imageUseCase.UpdateProductImage(
		c.Request.Context(),
		id,
		req.VariantID,
		req.DisplayOrder,
		req.IsPrimary,
	)

	if err != nil {
		h.handleProductImageError(
			c,
			err,
			"failed to update product image",
		)
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product image updated successfully",
		Data:    response.ToProductImageResponse(image),
	})
}
func (h *ProductImageHandler) DeleteProductImage(c *gin.Context) {
	id, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid image id",
		})
		return
	}
	err = h.imageUseCase.DeleteProductImage(
		c.Request.Context(),
		id,
	)
	if err != nil {
		h.handleProductImageError(
			c,
			err,
			"failed to delete product image",
		)
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product image deleted successfully",
	})
}
