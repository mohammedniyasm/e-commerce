package handler

import (
	"net/http"
	"strconv"
	"time"

	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardUseCase interfaces.DashboardUseCase
}

func NewDashboardHandler(
	dashboardUseCase interfaces.DashboardUseCase,
) *DashboardHandler {
	return &DashboardHandler{
		dashboardUseCase: dashboardUseCase,
	}
}

func (h *DashboardHandler) GetOverview(c *gin.Context) {
	result, err := h.dashboardUseCase.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch dashboard overview",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "dashboard overview fetched successfully",
		Data:    result,
	})
}
func (h *DashboardHandler) GetSales(c *gin.Context) {
	groupBy := c.DefaultQuery("group_by", "day")
	filter := models.DashboardFilter{
		GroupBy: groupBy,
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate != "" {
		parsed, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid start_date format; use YYYY-MM-DD",
			})
			return
		}
		filter.StartDate = parsed
	}
	if endDate != "" {
		parsed, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid end_date format; use YYYY-MM-DD",
			})
			return
		}
		filter.EndDate = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	result, err := h.dashboardUseCase.GetSales(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "failed to fetch sales data",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "sales data fetched successfully",
		Data:    result,
	})
}
func (h *DashboardHandler) GetOrderStats(c *gin.Context) {
	result, err := h.dashboardUseCase.GetOrderStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch order statistics",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "order statistics fetched successfully",
		Data:    result,
	})
}
func (h *DashboardHandler) GetCustomerStats(c *gin.Context) {
	result, err := h.dashboardUseCase.GetCustomerStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch customer statistics",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "customer statistics fetched successfully",
		Data:    result,
	})
}
func (h *DashboardHandler) GetTopProducts(c *gin.Context) {
	limit, err := parseDashboardLimit(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid limit",
		})
		return
	}
	result, err := h.dashboardUseCase.GetTopProducts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch top products",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "top products fetched successfully",
		Data:    result,
	})
}
func (h *DashboardHandler) GetTopCategories(c *gin.Context) {
	limit, err := parseDashboardLimit(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid limit",
		})
		return
	}
	result, err := h.dashboardUseCase.GetTopCategories(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch top categories",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "top categories fetched successfully",
		Data:    result,
	})
}
func (h *DashboardHandler) GetTopBrands(c *gin.Context) {
	limit, err := parseDashboardLimit(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid limit",
		})
		return
	}
	result, err := h.dashboardUseCase.GetTopBrands(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch top brands",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "top brands fetched successfully",
		Data:    result,
	})
}
func parseDashboardLimit(c *gin.Context) (int, error) {
	value := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 || limit > 100 {
		return 0, strconv.ErrSyntax
	}
	return limit, nil
}
