package handler

import (
	"errors"
	"net/http"
	"strconv"

	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

type ProductVariantHandler struct {
	variantUseCase interfaces.ProductVariantUseCase
}

func NewProductVariantHandler(
	variantUseCase interfaces.ProductVariantUseCase,
) *ProductVariantHandler {
	return &ProductVariantHandler{
		variantUseCase: variantUseCase,
	}
}

func (h *ProductVariantHandler) CreateVariant(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"),10,64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}
	var req request.CreateProductVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	variant, err := h.variantUseCase.CreateVariant(
		c.Request.Context(),
		productID,
		req.Size,
		req.Color,
		req.SKU,
		req.MRP,
		req.SellingPrice,
		req.Stock,
	)
	if err != nil {
		h.handleVariantError(c, err, "failed to create product variant")
		return
	}

	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "product variant created successfully",
		Data:    response.ToProductVariantResponse(variant),
	})
}
func (h *ProductVariantHandler) GetVariant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid variant id",
		})
		return
	}
	variant, err := h.variantUseCase.GetVariant(
		c.Request.Context(),
		id,
	)
	if err != nil {
		h.handleVariantError(c, err, "failed to fetch product variant")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product variant fetched successfully",
		Data:    response.ToProductVariantResponse(variant),
	})
}
func (h *ProductVariantHandler) UpdateVariant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid variant id",
		})
		return
	}
	var req request.UpdateProductVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	variant, err := h.variantUseCase.UpdateVariant(
		c.Request.Context(),
		id,
		req.Size,
		req.Color,
		req.SKU,
		req.MRP,
		req.SellingPrice,
		req.Stock,
		req.IsActive,
	)
	if err != nil {
		h.handleVariantError(c, err, "failed to update product variant")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product variant updated successfully",
		Data:    response.ToProductVariantResponse(variant),
	})
}
func (h *ProductVariantHandler) DeleteVariant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid variant id",
		})
		return
	}
	err = h.variantUseCase.DeleteVariant(
		c.Request.Context(),
		id,
	)
	if err != nil {
		h.handleVariantError(c, err, "failed to delete product variant")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product variant deleted successfully",
	})
}
func (h *ProductVariantHandler) ListVariantsByProductID(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"),10,64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}
	variants, err := h.variantUseCase.ListVariantsByProductID(
		c.Request.Context(),
		productID,
	)
	if err != nil {
		h.handleVariantError(
			c,
			err,
			"failed to fetch product variants",
		)
		return
	}
	variantResponses := make(
		[]response.ProductVariantResponse,
		0,
		len(variants),
	)
	for i := range variants {
		variantResponses = append(
			variantResponses,
			response.ToProductVariantResponse(&variants[i]),
		)
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product variants fetched successfully",
		Data:    variantResponses,
	})
}
func (h *ProductVariantHandler) handleVariantError(c *gin.Context,err error,message string) {

	switch {
	case errors.Is(err, domainerrors.ErrInvalidVariantID):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid variant id",
		})

	case errors.Is(err, domainerrors.ErrInvalidProductID):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})

	case errors.Is(err, domainerrors.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "product or variant not found",
		})

	case errors.Is(err, domainerrors.ErrInvalidVariantSKU):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "variant SKU is required",
		})

	case errors.Is(err, domainerrors.ErrVariantSKUAlreadyExists):
		c.JSON(http.StatusConflict, response.APIResponse{
			Success: false,
			Message: "variant SKU already exists",
		})

	case errors.Is(err, domainerrors.ErrVariantAlreadyExists):
		c.JSON(http.StatusConflict, response.APIResponse{
			Success: false,
			Message: "variant with the same size and color already exists",
		})

	case errors.Is(err, domainerrors.ErrInvalidMRP):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "MRP must be greater than zero",
		})

	case errors.Is(err, domainerrors.ErrInvalidSellingPrice):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "selling price must be greater than zero",
		})

	case errors.Is(err, domainerrors.ErrSellingPriceGreaterThanMRP):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "selling price cannot be greater than MRP",
		})

	case errors.Is(err, domainerrors.ErrInvalidStock):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "stock cannot be negative",
		})

	default:
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
	}
}
