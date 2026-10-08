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

type ProductHandler struct {
	productUseCase interfaces.ProductUseCase
}

func NewProductHandler(
	productUseCase interfaces.ProductUseCase,
) *ProductHandler {
	return &ProductHandler{
		productUseCase: productUseCase,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {

	var req request.CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	product, err := h.productUseCase.CreateProduct(
		c.Request.Context(),
		req.Name,
		req.ShortDescription,
		req.Description,
		req.CategoryID,
		req.BrandID,
	)

	if err != nil {

		if errors.Is(err, domainerrors.ErrInvalidProductName) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "product name is required",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrInvalidCategoryID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid category id",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrInvalidBrandID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid brand id",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "category or brand not found",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrCategoryInactive) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "category is inactive",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrBrandInactive) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "brand is inactive",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrProductSlugExists) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product slug already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create product",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "product created successfully",
		Data:    response.ToProductResponse(product),
	})
}
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}

	product, err := h.productUseCase.GetProduct(
		c.Request.Context(),
		uint(id),
	)

	if err != nil {

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch product",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product fetched successfully",
		Data:    response.ToProductResponse(product),
	})
}
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product slug",
		})
		return
	}
	product, err := h.productUseCase.GetProductBySlug(
		c.Request.Context(),
		slug,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidProductSlug) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid product slug",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch product",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product fetched successfully",
		Data:    response.ToProductResponse(product),
	})
}
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}
	var req request.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	product, err := h.productUseCase.UpdateProduct(
		c.Request.Context(),
		uint(id),
		req.Name,
		req.ShortDescription,
		req.Description,
		req.CategoryID,
		req.BrandID,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidProductName) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "product name is required",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidCategoryID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid category id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidBrandID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid brand id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product, category, or brand not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrCategoryInactive) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "category is inactive",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrBrandInactive) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "brand is inactive",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrProductSlugExists) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product slug already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to update product",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product updated successfully",
		Data:    response.ToProductResponse(product),
	})
}
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}
	err = h.productUseCase.DeleteProduct(
		c.Request.Context(),
		uint(id),
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to delete product",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product deleted successfully",
	})
}
func (h *ProductHandler) ListProducts(c *gin.Context) {
	var req request.ListProductRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid query parameters",
			Error:   err.Error(),
		})
		return
	}
	products, total, err := h.productUseCase.ListProducts(
		c.Request.Context(),
		req.Search,
		req.Page,
		req.Limit,
		req.IsActive,
		req.IsListed,
		req.CategoryID,
		req.BrandID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch products",
			Error:   err.Error(),
		})
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(
			(total + int64(req.Limit) - 1) /
				int64(req.Limit),
		)
	}
	productResponses := make(
		[]response.ProductResponse,
		0,
		len(products),
	)
	for i := range products {
		productResponses = append(
			productResponses,
			response.ToProductResponse(&products[i]),
		)
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "products fetched successfully",
		Data: response.ProductListResponse{
			Products:   productResponses,
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *ProductHandler) ListDeletedProducts(c *gin.Context) {
	var req request.ListProductRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid query parameters",
			Error:   err.Error(),
		})
		return
	}
	products, total, err := h.productUseCase.ListDeletedProducts(
		c.Request.Context(),
		req.Search,
		req.Page,
		req.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch deleted products",
			Error:   err.Error(),
		})
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}

	if req.Limit < 1 {
		req.Limit = 10
	}

	totalPages := 0

	if total > 0 {
		totalPages = int(
			(total + int64(req.Limit) - 1) /
				int64(req.Limit),
		)
	}

	productResponses := make(
		[]response.ProductResponse,
		0,
		len(products),
	)

	for i := range products {
		productResponses = append(
			productResponses,
			response.ToProductResponse(&products[i]),
		)
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "deleted products fetched successfully",
		Data: response.ProductListResponse{
			Products:   productResponses,
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *ProductHandler) RestoreProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}
	err = h.productUseCase.RestoreProduct(
		c.Request.Context(),
		uint(id),
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "deleted product not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrCategoryInactive) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "product category is inactive",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrBrandInactive) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "product brand is inactive",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to restore product",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product restored successfully",
	})
}
func (h *ProductHandler) ToggleProductActive(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}

	err = h.productUseCase.ToggleProductActive(
		c.Request.Context(),
		uint(id),
	)

	if err != nil {

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to toggle product active status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product active status updated successfully",
	})
}
func (h *ProductHandler) ToggleProductListed(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}

	err = h.productUseCase.ToggleProductListed(
		c.Request.Context(),
		uint(id),
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to toggle product listed status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product listed status updated successfully",
	})
}
