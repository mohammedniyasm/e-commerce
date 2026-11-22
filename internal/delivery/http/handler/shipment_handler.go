package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ShipmentHandler struct {
	shipmentUseCase interfaces.ShipmentUseCase
	log             *slog.Logger
}

func NewShipmentHandler(
	shipmentUseCase interfaces.ShipmentUseCase,
	log *slog.Logger,
) *ShipmentHandler {
	return &ShipmentHandler{
		shipmentUseCase: shipmentUseCase,
		log:             log,
	}
}
func (h *ShipmentHandler) ListShipments(c *gin.Context) {
	var req request.ListShipmentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid shipment query parameters",
		})
		return
	}

	filter, err := parseShipmentListFilter(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Sort == "" {
		filter.Sort = "newest"
	}

	shipments, total, err := h.shipmentUseCase.ListShipments(c.Request.Context(), filter)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidInput),
			errors.Is(err, domainerrors.ErrInvalidOrderID),
			errors.Is(err, domainerrors.ErrInvalidShippingStatus):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "Invalid shipment filters",
			})
		default:
			h.log.ErrorContext(
				c.Request.Context(),
				"failed to list shipments",
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "Unable to retrieve shipments",
			})
		}
		return
	}

	items := make([]response.ShipmentResponse, 0, len(shipments))
	for i := range shipments {
		items = append(items, toShipmentResponse(&shipments[i]))
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(filter.Limit) - 1) / int64(filter.Limit)
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Shipments retrieved successfully",
		Data: response.ShipmentListResponse{
			Shipments:  items,
			TotalCount: total,
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalPages: totalPages,
		},
	})
}
func (h *ShipmentHandler) CreateShipment(c *gin.Context) {
	var req request.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid shipment request",
		})
		return
	}
	if req.OrderID == uuid.Nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	shipment := &models.Shipment{
		OrderID:        req.OrderID,
		CourierName:    req.CourierName,
		TrackingNumber: req.TrackingNumber,
	}
	if err := h.shipmentUseCase.CreateShipment(
		c.Request.Context(),
		shipment,
	); err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidShipment):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "Invalid shipment",
			})
		case errors.Is(err, domainerrors.ErrInvalidOrderID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "Invalid order ID",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "Order not found",
			})
		default:
			h.log.ErrorContext(
				c.Request.Context(),
				"failed to create shipment",
				"order_id", req.OrderID,
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "Failed to create shipment",
			})
		}
		return
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "Shipment created successfully",
		Data:    toShipmentResponse(shipment),
	})
}
func (h *ShipmentHandler) GetShipmentByID(c *gin.Context) {
	shipmentID, ok := parseShipmentID(c)
	if !ok {
		return
	}
	shipment, err := h.shipmentUseCase.GetShipmentByID(
		c.Request.Context(),
		shipmentID,
	)
	if err != nil {
		h.handleShipmentError(c, err, "Failed to retrieve shipment")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Shipment retrieved successfully",
		Data:    toShipmentResponse(shipment),
	})
}
func (h *ShipmentHandler) GetShipmentByOrderID(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("order_id"))
	if err != nil || orderID == uuid.Nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	shipment, err := h.shipmentUseCase.GetShipmentByOrderID(
		c.Request.Context(),
		orderID,
	)
	if err != nil {
		h.handleShipmentError(c, err, "Failed to retrieve shipment")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Shipment retrieved successfully",
		Data:    toShipmentResponse(shipment),
	})
}
func (h *ShipmentHandler) UpdateShipment(c *gin.Context) {
	shipmentID, ok := parseShipmentID(c)
	if !ok {
		return
	}
	var req request.UpdateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid shipment update request",
		})
		return
	}
	shipment, err := h.shipmentUseCase.GetShipmentByID(
		c.Request.Context(),
		shipmentID,
	)
	if err != nil {
		h.handleShipmentError(c, err, "Failed to retrieve shipment")
		return
	}
	if req.CourierName != nil {
		shipment.CourierName = req.CourierName
	}
	if req.TrackingNumber != nil {
		shipment.TrackingNumber = req.TrackingNumber
	}
	if err := h.shipmentUseCase.UpdateShipment(
		c.Request.Context(),
		shipment,
	); err != nil {
		h.handleShipmentError(c, err, "Failed to update shipment")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Shipment updated successfully",
		Data:    toShipmentResponse(shipment),
	})
}
func (h *ShipmentHandler) UpdateShippingStatus(c *gin.Context) {
	shipmentID, ok := parseShipmentID(c)
	if !ok {
		return
	}
	var req request.UpdateShippingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid shipping status request",
		})
		return
	}
	if err := h.shipmentUseCase.UpdateShippingStatus(
		c.Request.Context(),
		shipmentID,
		req.Status,
		req.Description,
		req.Location,
	); err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidShipmentID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "Invalid shipment ID",
			})
		case errors.Is(err, domainerrors.ErrInvalidShippingStatus):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "Invalid shipping status",
			})
		case errors.Is(err, domainerrors.ErrInvalidShippingTransition):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "Invalid shipping status transition",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "Shipment not found",
			})
		default:
			h.log.ErrorContext(
				c.Request.Context(),
				"failed to update shipping status",
				"shipment_id", shipmentID,
				"status", req.Status,
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "Failed to update shipping status",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Shipping status updated successfully",
		Data: response.ShipmentMessageResponse{
			ShipmentID: shipmentID,
			Message:    "Shipment status updated successfully",
		},
	})
}
func (h *ShipmentHandler) GetTrackingHistory(c *gin.Context) {
	shipmentID, ok := parseShipmentID(c)
	if !ok {
		return
	}
	history, err := h.shipmentUseCase.GetTrackingHistory(
		c.Request.Context(),
		shipmentID,
	)
	if err != nil {
		h.handleShipmentError(c, err, "Failed to retrieve tracking history")
		return
	}
	historyResponses := make(
		[]response.ShipmentTrackingHistoryResponse,
		0,
		len(history),
	)
	for _, item := range history {
		historyResponses = append(
			historyResponses,
			response.ShipmentTrackingHistoryResponse{
				ID:             item.ID,
				ShipmentID:     item.ShipmentID,
				ShippingStatus: item.ShippingStatus,
				Description:    item.Description,
				Location:       item.Location,
				CreatedAt:      item.CreatedAt,
			},
		)
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Shipment tracking history retrieved successfully",
		Data: response.ShipmentTrackingHistoryListResponse{
			ShipmentID: shipmentID,
			History:    historyResponses,
			Total:      len(historyResponses),
		},
	})
}
func (h *ShipmentHandler) GetCustomerShipment(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("order_id"))
	if err != nil || orderID == uuid.Nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	authenticatedUserID, ok := getCartUserID(c)
	if !ok || authenticatedUserID == 0 {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Invalid authentication context",
		})
		return
	}
	shipment, err := h.shipmentUseCase.GetCustomerShipment(
		c.Request.Context(),
		uint(authenticatedUserID),
		orderID,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"Failed to retrieve customer shipment",
			"order_id", orderID,
			"user_id", authenticatedUserID,
			"error", err,
		)
		h.handleShipmentError(c, err, "")
		return
	}
	data := response.CustomerShipmentResponse{
		ID:             shipment.ID,
		OrderID:        shipment.OrderID,
		CourierName:    shipment.CourierName,
		TrackingNumber: shipment.TrackingNumber,
		ShippingStatus: shipment.ShippingStatus,
		ShippedAt:      shipment.ShippedAt,
		DeliveredAt:    shipment.DeliveredAt,
		CreatedAt:      shipment.CreatedAt,
		UpdatedAt:      shipment.UpdatedAt,
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Shipment retrieved successfully",
		Data:    data,
	})
}
func (h *ShipmentHandler) GetCustomerTrackingHistory(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("order_id"))
	if err != nil || orderID == uuid.Nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	authenticatedUserID, ok := getCartUserID(c)
	if !ok || authenticatedUserID == 0 {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Invalid authentication context",
		})
		return
	}
	history, err := h.shipmentUseCase.GetCustomerTrackingHistory(
		c.Request.Context(),
		uint(authenticatedUserID),
		orderID,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"Failed to retrieve customer shipment history",
			"order_id", orderID,
			"user_id", authenticatedUserID,
			"error", err,
		)
		h.handleShipmentError(c, err, "")
		return
	}
	shipment, err := h.shipmentUseCase.GetCustomerShipment(
		c.Request.Context(),
		uint(authenticatedUserID),
		orderID,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"Failed to retrieve shipment for tracking history response",
			"order_id", orderID,
			"user_id", authenticatedUserID,
			"error", err,
		)
		h.handleShipmentError(c, err, "")
		return
	}
	items := make(
		[]response.CustomerShipmentHistoryItemResponse,
		0,
		len(history),
	)
	for _, item := range history {
		items = append(items, response.CustomerShipmentHistoryItemResponse{
			ID:             item.ID,
			ShippingStatus: item.ShippingStatus,
			Description:    item.Description,
			Location:       item.Location,
			CreatedAt:      item.CreatedAt,
		})
	}
	data := response.CustomerShipmentHistoryResponse{
		ShipmentID: shipment.ID,
		History:    items,
		Total:      len(items),
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Shipment tracking history retrieved successfully",
		Data:    data,
	})
}
func parseShipmentID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(
		c.Param("shipment_id"),
		10,
		32,
	)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid shipment ID",
		})
		return 0, false
	}

	return uint(id), true
}
func toShipmentResponse(shipment *models.Shipment) response.ShipmentResponse {
	history := make(
		[]response.ShipmentTrackingHistoryResponse,
		0,
		len(shipment.History),
	)
	for _, item := range shipment.History {
		history = append(history, response.ShipmentTrackingHistoryResponse{
			ID:             item.ID,
			ShipmentID:     item.ShipmentID,
			ShippingStatus: item.ShippingStatus,
			Description:    item.Description,
			Location:       item.Location,
			CreatedAt:      item.CreatedAt,
		})
	}
	return response.ShipmentResponse{
		ID:              shipment.ID,
		OrderID:         shipment.OrderID,
		OrderNumber:     shipment.Order.OrderNumber,
		Type:            shipment.Type,
		ReturnRequestID: shipment.ReturnRequestID,
		CourierName:     shipment.CourierName,
		TrackingNumber:  shipment.TrackingNumber,
		ShippingStatus:  shipment.ShippingStatus,
		ShippedAt:       shipment.ShippedAt,
		DeliveredAt:     shipment.DeliveredAt,
		CreatedAt:       shipment.CreatedAt,
		UpdatedAt:       shipment.UpdatedAt,
		History:         history,
	}
}

func parseShipmentListFilter(req request.ListShipmentsRequest) (models.ShipmentListFilter, error) {
	filter := models.ShipmentListFilter{
		Search:      strings.TrimSpace(req.Search),
		CourierName: strings.TrimSpace(req.CourierName),
		Sort:        strings.TrimSpace(req.Sort),
		Limit:       req.Limit,
		Page:        req.Page,
	}

	if status := strings.TrimSpace(req.ShippingStatus); status != "" {
		switch status {
		case "pending", "packed", "ready_to_ship", "in_transit", "out_for_delivery",
			"delivered", "failed", "returned", "cancelled":
		default:
			return models.ShipmentListFilter{}, errors.New("invalid shipping_status")
		}
		filter.ShippingStatus = &status
	}
	if shipmentType := strings.TrimSpace(req.Type); shipmentType != "" {
		value := models.ShipmentType(shipmentType)
		switch value {
		case models.ShipmentTypeDelivery, models.ShipmentTypeReturn, models.ShipmentTypeReturnToCustomer:
		default:
			return models.ShipmentListFilter{}, errors.New("invalid type")
		}
		filter.Type = &value
	}
	switch filter.Sort {
	case "", "newest", "oldest", "updated_desc", "updated_asc", "status_asc", "status_desc":
	default:
		return models.ShipmentListFilter{}, errors.New("invalid sort")
	}
	if rawOrderID := strings.TrimSpace(req.OrderID); rawOrderID != "" {
		orderID, err := uuid.Parse(rawOrderID)
		if err != nil || orderID == uuid.Nil {
			return models.ShipmentListFilter{}, errors.New("invalid order_id")
		}
		filter.OrderID = &orderID
	}
	if rawReturnRequestID := strings.TrimSpace(req.ReturnRequestID); rawReturnRequestID != "" {
		parsedID, err := strconv.ParseUint(rawReturnRequestID, 10, 32)
		if err != nil || parsedID == 0 {
			return models.ShipmentListFilter{}, errors.New("invalid return_request_id")
		}
		id := uint(parsedID)
		filter.ReturnRequestID = &id
	}
	createdFrom, err := parseShipmentDateFilter(req.CreatedFrom)
	if err != nil {
		return models.ShipmentListFilter{}, errors.New("created_from must use YYYY-MM-DD format")
	}
	createdTo, err := parseShipmentDateFilter(req.CreatedTo)
	if err != nil {
		return models.ShipmentListFilter{}, errors.New("created_to must use YYYY-MM-DD format")
	}
	filter.CreatedFrom = createdFrom
	filter.CreatedTo = createdTo
	if createdFrom != nil && createdTo != nil && createdFrom.After(*createdTo) {
		return models.ShipmentListFilter{}, errors.New("created_from cannot be after created_to")
	}
	return filter, nil
}

func parseShipmentDateFilter(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
func (h *ShipmentHandler) handleShipmentError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, domainerrors.ErrInvalidShipmentID),
		errors.Is(err, domainerrors.ErrInvalidOrderID):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid shipment or order ID",
		})

	case errors.Is(err, domainerrors.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "Shipment not found",
		})

	default:
		h.log.ErrorContext(
			c.Request.Context(),
			"shipment request failed",
			"error", err,
		)

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: message,
		})
	}
}
