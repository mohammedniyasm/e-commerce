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

type BrandHandler struct {
	brandUseCase interfaces.BrandUseCase
}

func NewBrandHandler(
	brandUseCase interfaces.BrandUseCase,
) *BrandHandler {
	return &BrandHandler{
		brandUseCase: brandUseCase,
	}
}

func (h *BrandHandler) CreateBrand(c *gin.Context) {

	var req request.CreateBrandRequest

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

	brand, err := h.brandUseCase.CreateBrand(
		c.Request.Context(),
		req.Name,
		description,
	)
	if err != nil {

		if errors.Is(err, domainerrors.ErrInvalidBrandName) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "brand name is required",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create brand",
			Error:   err.Error(),
		})
		return
	}
	fileHeader, err := c.FormFile("logo")
	if err == nil {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to open brand logo",
			})
			return
		}
		defer file.Close()

		brand, err = h.brandUseCase.UploadBrandLogo(
			c.Request.Context(),
			int64(brand.ID),
			file,
			fileHeader.Size,
			fileHeader.Header.Get("Content-Type"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "brand created but logo upload failed",
				Error:   err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "brand created successfully",
		Data:    response.ToBrandResponse(brand),
	})
}
func (h *BrandHandler) GetBrand(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid brand id",
		})
		return
	}

	brand, err := h.brandUseCase.GetBrand(
		c.Request.Context(),
		id,
	)
	if err != nil {

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "brand not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch brand",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "brand fetched successfully",
		Data:    response.ToBrandResponse(brand),
	})
}
func (h *BrandHandler) UpdateBrand(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid brand id",
		})
		return
	}

	var req request.UpdateBrandRequest

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

	brand, err := h.brandUseCase.UpdateBrand(
		c.Request.Context(),
		id,
		req.Name,
		description,
	)
	if err != nil {

		if errors.Is(err, domainerrors.ErrInvalidBrandName) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "brand name is required",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "brand not found",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrBrandNameAlreadyExists) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "brand name already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to update brand",
			Error:   err.Error(),
		})
		return
	}
	fileHeader, err := c.FormFile("logo")
	if err == nil {

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to open brand logo",
			})
			return
		}
		defer file.Close()

		brand, err = h.brandUseCase.UploadBrandLogo(
			c.Request.Context(),
			id,
			file,
			fileHeader.Size,
			fileHeader.Header.Get("Content-Type"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "brand updated but logo upload failed",
				Error:   err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "brand updated successfully",
		Data:    response.ToBrandResponse(brand),
	})
}
func (h *BrandHandler) DeleteBrand(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid brand id",
		})
		return
	}

	err = h.brandUseCase.DeleteBrand(
		c.Request.Context(),
		id,
	)
	if err != nil {

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "brand not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to delete brand",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "brand deleted successfully",
	})
}
func (h *BrandHandler) ListBrands(c *gin.Context) {

	var req request.ListBrandRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	brands, total, err := h.brandUseCase.ListBrands(
		c.Request.Context(),
		req.Search,
		req.Page,
		req.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch brands",
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

	brandResponses := make(
		[]response.BrandResponse,
		0,
		len(brands),
	)

	for i := range brands {
		brandResponses = append(
			brandResponses,
			response.ToBrandResponse(&brands[i]),
		)
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "brands fetched successfully",
		Data: response.BrandListResponse{
			Brands:     brandResponses,
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *BrandHandler) ListDeletedBrands(c *gin.Context) {

	var req request.ListBrandRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	brands, total, err := h.brandUseCase.ListDeletedBrands(
		c.Request.Context(),
		req.Search,
		req.Page,
		req.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch deleted brands",
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

	brandResponses := make(
		[]response.BrandResponse,
		0,
		len(brands),
	)

	for i := range brands {
		brandResponses = append(
			brandResponses,
			response.ToBrandResponse(&brands[i]),
		)
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "deleted brands fetched successfully",
		Data: response.BrandListResponse{
			Brands:     brandResponses,
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *BrandHandler) RestoreBrand(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid brand id",
		})
		return
	}

	err = h.brandUseCase.RestoreBrand(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "deleted brand not found",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrRestoreBrandAlreadyExists) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "brand cannot be restored because the name already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to restore brand",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "brand restored successfully",
	})
}
func (h *BrandHandler) ToggleBrandActive(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid brand id",
		})
		return
	}

	err = h.brandUseCase.ToggleBrandActive(
		c.Request.Context(),
		id,
	)
	if err != nil {

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "brand not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to toggle brand status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "brand status updated successfully",
	})
}
func (h *BrandHandler) DeleteBrandLogo(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid brand id",
		})
		return
	}

	err = h.brandUseCase.DeleteBrandLogo(
		c.Request.Context(),
		id,
	)
	if err != nil {

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "brand not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to delete brand logo",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "brand logo deleted successfully",
	})
}
func (h *BrandHandler) UpdateBrandLogo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid brand id",
		})
		return
	}
	fileHeader, err := c.FormFile("logo")
	if err != nil {
		c.JSON(404, response.APIResponse{
			Success: false,
			Message: "brand logo is required",
			Error:   err.Error(),
		})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to open brand logo",
		})
		return
	}
	defer file.Close()
	brand, err := h.brandUseCase.UploadBrandLogo(
		c.Request.Context(),
		id,
		file,
		fileHeader.Size,
		fileHeader.Header.Get("Content-Type"),
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "brand not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidBrandLogo) ||
			errors.Is(err, domainerrors.ErrInvalidBrandLogoSize) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to update brand logo",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "brand logo updated successfully",
		Data:    response.ToBrandResponse(brand),
	})
}
