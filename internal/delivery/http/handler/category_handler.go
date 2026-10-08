package handler

import (
	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryUseCase interfaces.CategoryUseCase
}

func NewCategoryHandler(
	categoryUseCase interfaces.CategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		categoryUseCase: categoryUseCase,
	}
}
func (h *CategoryHandler) CreateCategory(c *gin.Context) {

	var req request.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	var description string
	if req.Description != nil {
		description = *req.Description
	}

	category, err := h.categoryUseCase.CreateCategory(
		c.Request.Context(),
		req.Name,
		description,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCategoryName) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "category name is required",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "category created successfully",
		Data:    response.ToCategoryResponse(category),
	})
}
func (h *CategoryHandler) GetCategory(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid category id",
		})
		return
	}

	category, err := h.categoryUseCase.GetCategory(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "category fetched successfully",
		Data:    response.ToCategoryResponse(category),
	})
}
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid category id",
		})
		return
	}

	var req request.UpdateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	var description string
	if req.Description != nil {
		description = *req.Description
	}

	category, err := h.categoryUseCase.UpdateCategory(
		c.Request.Context(),
		id,
		req.Name,
		description,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCategoryName) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "category name is required",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to update category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "category updated successfully",
		Data:    response.ToCategoryResponse(category),
	})
}
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid category id",
		})
		return
	}

	err = h.categoryUseCase.DeleteCategory(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "category not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrCategoryHasActiveProducts) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "category cannot be deleted because it has active products",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to delete category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "category deleted successfully",
	})
}
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	var req request.ListCategoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	categories, total, err := h.categoryUseCase.ListCategories(
		c.Request.Context(),
		req.Search,
		req.Page,
		req.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch categories",
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
	categoryResponses := make([]response.CategoryResponse, 0, len(categories))
	for i := range categories {
		categoryResponses = append(categoryResponses, response.ToCategoryResponse(&categories[i]))
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "categories fetched successfully",
		Data: response.CategoryListResponse{
			Categories: categoryResponses,
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *CategoryHandler) ListDeletedCategories(c *gin.Context) {

	var req request.ListCategoryRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	categories, total, err := h.categoryUseCase.ListDeletedCategories(
		c.Request.Context(),
		req.Search,
		req.Page,
		req.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch deleted categories",
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

	categoryResponses := make(
		[]response.CategoryResponse,
		0,
		len(categories),
	)

	for i := range categories {
		categoryResponses = append(
			categoryResponses,
			response.ToCategoryResponse(&categories[i]),
		)
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "deleted categories fetched successfully",
		Data: response.CategoryListResponse{
			Categories: categoryResponses,
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *CategoryHandler) RestoreCategory(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid category id",
		})
		return
	}

	err = h.categoryUseCase.RestoreCategory(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "deleted category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to restore category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "category restored successfully",
	})
}
func (h *CategoryHandler) ToggleCategoryActive(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid category id",
		})
		return
	}

	err = h.categoryUseCase.ToggleCategoryActive(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to toggle category status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "category status updated successfully",
	})
}
