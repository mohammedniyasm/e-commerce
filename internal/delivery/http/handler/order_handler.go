package handler

import (
	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderUseCase interfaces.OrderUseCase
	log          *slog.Logger
}

func NewOrderHandler(
	orderUseCase interfaces.OrderUseCase,
	log *slog.Logger,
) *OrderHandler {
	return &OrderHandler{
		orderUseCase: orderUseCase,
		log:          log,
	}
}
func getAuthenticatedUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch id := value.(type) {
	case int64:
		return uint(id), id > 0

	case int:
		return uint(id), id > 0

	case uint:
		if id == 0 {
			return 0, false
		}
		return id, true

	case string:
		parsedID, err := strconv.ParseInt(id, 10, 64)
		if err != nil || parsedID <= 0 {
			return 0, false
		}
		return uint(parsedID), true

	default:
		return 0, false
	}
}
func parseOrderID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.Param("order_id"))
}
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}
	var req request.PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid request",
			// "error":   err.Error(),
		})
		return
	}
	paymentMethod := strings.TrimSpace(req.PaymentMethod)
	if paymentMethod == "" {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Payment method is required",
		})
		return
	}
	var couponCode *string
	if code := strings.TrimSpace(req.CouponCode); code != "" {
		couponCode = &code
	}
	order, err := h.orderUseCase.PlaceOrder(
		c.Request.Context(),
		userID,
		req.CartItemIDs,
		req.AddressID,
		paymentMethod,
		couponCode,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to place order",
			"user_id", userID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Unable to place order",
		})
		return
	}
	start, end := calculateEstimatedDelivery(order.PlacedAt)
	res := response.PlaceOrderResponse{
		OrderID:                order.ID,
		OrderNumber:            order.OrderNumber,
		OrderStatus:            order.OrderStatus,
		PaymentStatus:          order.PaymentStatus,
		Subtotal:               order.Subtotal,
		DiscountAmount:         order.DiscountAmount,
		ShippingAmount:         order.ShippingAmount,
		TotalAmount:            order.TotalAmount,
		EstimatedDeliveryStart: start,
		EstimatedDeliveryEnd:   end,
		CreatedAt:              order.CreatedAt,
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "Order placed successfully",
		Data:    res,
	})
}
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}
	orderID, err := parseOrderID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	if err := h.orderUseCase.CancelOrder(
		c.Request.Context(),
		userID,
		orderID,
	); err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to cancel order",
			"user_id", userID,
			"order_id", orderID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Unable to cancel order",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Order cancelled successfully",
		Data: response.OrderActionResponse{
			OrderID: orderID,
			Message: "Order cancelled successfully",
		},
	})
}
func (h *OrderHandler) UpdateOrderAddress(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}
	orderID, err := uuid.Parse(c.Param("order_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	var req request.UpdateOrderAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}
	address := models.OrderAddress{
		Name:         req.Name,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: &req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.Pincode,
	}
	if err := h.orderUseCase.UpdateOrderAddress(
		c.Request.Context(),
		userID,
		orderID,
		address,
	); err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to update order address",
			"user_id", userID,
			"order_id", orderID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Unable to update order address",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Order address updated successfully",
		Data: response.OrderActionResponse{
			OrderID: orderID,
			Message: "Order address updated successfully",
		},
	})
}
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("order_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	var req request.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if err := h.orderUseCase.UpdateOrderStatus(
		c.Request.Context(),
		orderID,
		req.Status,
		req.Note,
	); err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to update order status",
			"order_id", orderID,
			"status", req.Status,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Unable to update order status",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Order status updated successfully",
		Data: response.OrderActionResponse{
			OrderID: orderID,
			Message: "Order status updated successfully",
		},
	})
}
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}
	orderID, err := parseOrderID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	order, err := h.orderUseCase.GetByID(
		c.Request.Context(),
		orderID,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to retrieve order",
			"user_id", userID,
			"order_id", orderID,
			"error", err,
		)
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "Order not found",
		})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "Order not found",
		})
		return
	}
	res := mapOrderToResponse(order)
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Order retrieved successfully",
		Data:    res,
	})
}
func (h *OrderHandler) GetOrderByNumber(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}
	orderNumber := strings.TrimSpace(c.Param("order_number"))
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Order number is required",
		})
		return
	}
	order, err := h.orderUseCase.GetByOrderNumber(
		c.Request.Context(),
		orderNumber,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to retrieve order by order number",
			"user_id", userID,
			"order_number", orderNumber,
			"error", err,
		)

		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "Order not found",
		})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "Order not found",
		})
		return
	}
	res := mapOrderToResponse(order)
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Order retrieved successfully",
		Data:    res,
	})
}
func (h *OrderHandler) ListMyOrders(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}
	var req request.ListCustomerOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid query parameters",
		})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	orders, total, err := h.orderUseCase.ListByUserID(
		c.Request.Context(),
		userID,
		req.Limit,
		req.Page,
		req.Sort,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to list customer orders",
			"user_id", userID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "Unable to retrieve orders",
		})
		return
	}
	orderResponses := make([]response.OrderResponse, 0, len(orders))
	for i := range orders {
		orderResponses = append(
			orderResponses,
			mapOrderToResponse(&orders[i]),
		)
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(req.Limit) - 1) / int64(req.Limit)
	}
	result := response.OrderListResponse{
		Orders:     orderResponses,
		TotalCount: total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Orders retrieved successfully",
		Data:    result,
	})
}
func (h *OrderHandler) ListOrders(c *gin.Context) {
	var req request.ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid query parameters",
		})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	var userID *uint
	if rawUserID := strings.TrimSpace(c.Query("user_id")); rawUserID != "" {
		parsedID, err := strconv.ParseUint(rawUserID, 10, 32)
		if err != nil || parsedID == 0 {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "Invalid user ID",
			})
			return
		}
		id := uint(parsedID)
		userID = &id
	}
	orders, total, err := h.orderUseCase.List(
		c.Request.Context(),
		userID,
		strings.TrimSpace(req.Search),
		req.Limit,
		req.Page,
		req.Sort,
		req.OrderStatus,
		req.PaymentStatus,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to list orders",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "Unable to retrieve orders",
		})
		return
	}
	orderResponses := make([]response.OrderResponse, 0, len(orders))
	for i := range orders {
		orderResponses = append(
			orderResponses,
			mapOrderToResponse(&orders[i]),
		)
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(req.Limit) - 1) / int64(req.Limit)
	}
	result := response.OrderListResponse{
		Orders:     orderResponses,
		TotalCount: total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Orders retrieved successfully",
		Data:    result,
	})
}
func (h *OrderHandler) DownloadInvoice(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}
	orderID, err := parseOrderID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	pdfBytes, err := h.orderUseCase.GenerateInvoice(
		c.Request.Context(),
		userID,
		orderID,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to generate order invoice",
			"user_id", userID,
			"order_id", orderID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Unable to generate invoice",
		})
		return
	}
	filename := "invoice-" + orderID.String() + ".pdf"
	c.Header("Content-Type", "application/pdf")
	c.Header(
		"Content-Disposition",
		`attachment; filename="`+filename+`"`,
	)
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
func (h *OrderHandler) UpdateCODPaymentStatus(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("order_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	var req request.UpdateCODPaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}
	if err := h.orderUseCase.UpdateCODPaymentStatus(
		c.Request.Context(),
		orderID,
		req.Status,
	); err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to update COD payment status",
			"order_id", orderID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Unable to update COD payment status",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "COD payment status updated successfully",
	})
}
func shouldShowEstimatedDelivery(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "shipped", "out_for_delivery", "processing", "confirmed", "pending":
		return true
	default:
		return false
	}
}
func calculateEstimatedDelivery(placedAt time.Time) (*time.Time, *time.Time) {
	start := placedAt.AddDate(0, 0, 3)
	end := placedAt.AddDate(0, 0, 7)

	return &start, &end
}
func mapOrderToResponse(order *models.Order) response.OrderResponse {
	var estimatedStart, estimatedEnd *time.Time
	if shouldShowEstimatedDelivery(order.OrderStatus) {
		estimatedStart, estimatedEnd = calculateEstimatedDelivery(order.PlacedAt)
	}
	items := make([]response.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, response.OrderItemResponse{
			ID:          item.ID,
			VariantID:   item.VariantID,
			ProductName: item.ProductName,
			ImageURL:    item.ImageURL,
			VariantInfo: item.VariantInfo,
			UnitPrice:   item.UnitPrice,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		})
	}

	var address *response.OrderAddressResponse

	if order.Address != nil {
		address = &response.OrderAddressResponse{
			Name:         order.Address.Name,
			Phone:        order.Address.Phone,
			AddressLine1: order.Address.AddressLine1,
			City:         order.Address.City,
			State:        order.Address.State,
			Pincode:      order.Address.PostalCode,
		}

		if order.Address.AddressLine2 != nil {
			address.AddressLine2 = *order.Address.AddressLine2
		}
	}

	return response.OrderResponse{
		ID:                     order.ID,
		OrderNumber:            order.OrderNumber,
		OrderStatus:            order.OrderStatus,
		PaymentStatus:          order.PaymentStatus,
		Subtotal:               order.Subtotal,
		DiscountAmount:         order.DiscountAmount,
		ShippingAmount:         order.ShippingAmount,
		TotalAmount:            order.TotalAmount,
		EstimatedDeliveryStart: estimatedStart,
		EstimatedDeliveryEnd:   estimatedEnd,
		Items:                  items,
		Address:                address,
		PlacedAt:               order.PlacedAt,
		CreatedAt:              order.CreatedAt,
	}
}
