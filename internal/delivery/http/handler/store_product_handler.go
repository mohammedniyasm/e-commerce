package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

type StoreProductHandler struct {
	productUseCase interfaces.ProductUseCase
}

func NewStoreProductHandler(
	productUseCase interfaces.ProductUseCase,
) *StoreProductHandler {
	return &StoreProductHandler{
		productUseCase: productUseCase,
	}
}

func (h *StoreProductHandler) ListStoreProducts(c *gin.Context) {
	var req request.ListStoreProductsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}

	if req.Limit < 1 {
		req.Limit = 12
	}

	if req.Limit > 100 {
		req.Limit = 100
	}
	products, total, err := h.productUseCase.ListStoreProducts(
		c.Request.Context(),
		req.Search,
		req.CategoryID,
		req.BrandID,
		req.MinPrice,
		req.MaxPrice,
		req.Sort,
		req.Page,
		req.Limit,
	)

	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidCategoryID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid category id",
			})
			return

		case errors.Is(err, domainerrors.ErrInvalidBrandID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid brand id",
			})
			return

		case errors.Is(err, domainerrors.ErrInvalidPriceRange):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid price range",
			})
			return

		case errors.Is(err, domainerrors.ErrInvalidProductSort):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid product sort option",
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch products",
				Error:   err.Error(),
			})
			return
		}
	}

	productResponses := make(
		[]response.StoreProductResponse,
		0,
		len(products),
	)

	for i := range products {
		productResponses = append(
			productResponses,
			response.ToStoreProductResponse(&products[i]),
		)
	}
	totalPages := 0

	if total > 0 {
		totalPages = int(
			(total + int64(req.Limit) - 1) /
				int64(req.Limit),
		)
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "products fetched successfully",
		Data: response.StoreProductListResponse{
			Products:   productResponses,
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *StoreProductHandler) GetStoreProductByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid product id",
		})
		return
	}

	product, err := h.productUseCase.GetStoreProductByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidProductID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid product id",
			})

		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found or unavailable",
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch product details",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product details fetched successfully",
		Data:    response.ToStoreProductDetailResponse(product),
	})
}
func (h *StoreProductHandler) GetStoreProductBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.productUseCase.GetStoreProductBySlug(
		c.Request.Context(),
		slug,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidProductSlug):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid product slug",
			})

		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found or unavailable",
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch product details",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product details fetched successfully",
		Data:    response.ToStoreProductDetailResponse(product),
	})
}
func (h *StoreProductHandler) GetRelatedProducts(c *gin.Context) {
	// Parse the product ID from the URL.
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid product ID",
			"error":   "product ID must be a positive integer",
		})
		return
	}

	var query request.RelatedProductsQueryRequest

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid query parameters",
			"error":   err.Error(),
		})
		return
	}

	products, err := h.productUseCase.GetRelatedProducts(
		c.Request.Context(),
		productID,
		query.Limit,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "product not found",
				"error":   "product does not exist or is unavailable",
			})
			return
		}

		slog.Default().Error(
			"failed to retrieve related products",
			"product_id", productID,
			"error", err,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to retrieve related products",
			"error":   "internal server error",
		})
		return
	}

	data := response.ToRelatedProductsResponse(products)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "related products retrieved successfully",
		"data":    data,
	})
}
