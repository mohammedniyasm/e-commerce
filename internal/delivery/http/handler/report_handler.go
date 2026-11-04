package handler

import (
	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/infrastructure/reportexport"
	"ecommerce/internal/usecase/interfaces"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportUseCase interfaces.ReportUseCase
}

func NewReportHandler(
	reportUseCase interfaces.ReportUseCase,
) *ReportHandler {
	return &ReportHandler{
		reportUseCase: reportUseCase,
	}
}

func (h *ReportHandler) GetSalesReport(c *gin.Context) {
	filter, ok := parseReportFilter(c, true)
	if !ok {
		return
	}
	result, err := h.reportUseCase.GetSalesReport(
		c.Request.Context(),
		filter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to generate sales report",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "sales report generated successfully",
		Data:    result,
	})
}
func (h *ReportHandler) GetOrdersReport(c *gin.Context) {
	filter, ok := parseReportFilter(c, true)
	if !ok {
		return
	}
	result, err := h.reportUseCase.GetOrdersReport(
		c.Request.Context(),
		filter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to generate orders report",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "orders report generated successfully",
		Data:    result,
	})
}
func (h *ReportHandler) GetProductsReport(c *gin.Context) {
	filter, ok := parseReportFilter(c, false)
	if !ok {
		return
	}
	result, err := h.reportUseCase.GetProductsReport(
		c.Request.Context(),
		filter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to generate products report",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "products report generated successfully",
		Data:    result,
	})
}
func parseReportFilter(c *gin.Context, includeGroupBy bool) (models.ReportFilter, bool) {
	var filter models.ReportFilter
	startDate := strings.TrimSpace(c.Query("start_date"))
	endDate := strings.TrimSpace(c.Query("end_date"))
	if startDate != "" {
		parsed, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid start_date; use YYYY-MM-DD",
			})
			return filter, false
		}
		filter.StartDate = parsed
	}
	if endDate != "" {
		parsed, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid end_date; use YYYY-MM-DD",
			})
			return filter, false
		}

		// Include the entire end date.
		filter.EndDate = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	if !filter.StartDate.IsZero() &&
		!filter.EndDate.IsZero() &&
		filter.StartDate.After(filter.EndDate) {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "start_date cannot be after end_date",
		})
		return filter, false
	}
	if includeGroupBy {
		filter.GroupBy = strings.ToLower(strings.TrimSpace(c.DefaultQuery("group_by", "day")))
		switch filter.GroupBy {
		case "day", "week", "month", "year":
		default:
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid group_by; use day, week, month, or year",
			})
			return filter, false
		}
	}
	return filter, true
}
func (h *ReportHandler) ExportSalesPDF(c *gin.Context) {
	filter, ok := parseReportFilter(c, true)
	if !ok {
		return
	}
	report, err := h.reportUseCase.GetSalesReport(
		c.Request.Context(), filter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to generate sales report",
		})
		return
	}
	data, err := reportexport.GenerateSalesPDF(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create PDF",
		})
		return
	}
	c.Data(
		http.StatusOK,
		"application/pdf",
		data,
	)
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=sales-report-%s.pdf",
			time.Now().Format("20060102"),
		),
	)
}
func (h *ReportHandler) ExportSalesExcel(c *gin.Context) {
	filter, ok := parseReportFilter(c, true)
	if !ok {
		return
	}
	report, err := h.reportUseCase.GetSalesReport(
		c.Request.Context(), filter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to generate sales report",
		})
		return
	}
	data, err := reportexport.GenerateSalesExcel(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create Excel file",
		})
		return
	}
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=sales-report-%s.xlsx",
			time.Now().Format("20060102"),
		),
	)
	c.Data(
		http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		data,
	)
}
func (h *ReportHandler) ExportOrdersPDF(c *gin.Context) {
	filter, ok := parseReportFilter(c, true)
	if !ok {
		return
	}
	report, err := h.reportUseCase.GetOrdersReport(
		c.Request.Context(), filter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to generate orders report",
		})
		return
	}
	data, err := reportexport.GenerateOrdersPDF(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create PDF",
		})
		return
	}
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=orders-report-%s.pdf",
			time.Now().Format("20060102"),
		),
	)
	c.Data(http.StatusOK, "application/pdf", data)
}
func (h *ReportHandler) ExportOrdersExcel(c *gin.Context) {
	filter, ok := parseReportFilter(c, true)
	if !ok {
		return
	}
	report, err := h.reportUseCase.GetOrdersReport(
		c.Request.Context(), filter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to generate orders report",
		})
		return
	}
	data, err := reportexport.GenerateOrdersExcel(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create Excel file",
		})
		return
	}
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=orders-report-%s.xlsx",
			time.Now().Format("20060102"),
		),
	)
	c.Data(
		http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		data,
	)
}
func (h *ReportHandler) ExportProductsExcel(c *gin.Context) {
	filter, ok := parseReportFilter(c, false)
	if !ok {
		return
	}
	report, err := h.reportUseCase.GetProductsReport(
		c.Request.Context(), filter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to generate products report",
		})
		return
	}
	data, err := reportexport.GenerateProductsExcel(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create Excel file",
		})
		return
	}
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=products-report-%s.xlsx",
			time.Now().Format("20060102"),
		),
	)

	c.Data(
		http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		data,
	)
}
