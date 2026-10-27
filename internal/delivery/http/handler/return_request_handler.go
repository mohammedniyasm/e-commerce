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
	"github.com/google/uuid"
)

type ReturnRequestHandler struct {
	returnRequestUseCase interfaces.ReturnRequestUseCase
}

func NewReturnRequestHandler(
	returnRequestUseCase interfaces.ReturnRequestUseCase,
) *ReturnRequestHandler {
	return &ReturnRequestHandler{
		returnRequestUseCase: returnRequestUseCase,
	}
}
func (h *ReturnRequestHandler) Create(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	var req request.CreateReturnRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid order id",
		})
		return
	}
	returnRequest, err := h.returnRequestUseCase.Create(
		c.Request.Context(),
		uint(userID),
		orderID,
		req.OrderItemID,
		req.RequestType,
		req.RequestedVariantID,
		req.Quantity,
		req.Reason,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "order or order item not found",
			})
		case errors.Is(err, domainerrors.ErrInvalidOrderStatus):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "return or exchange is not available for this order",
			})
		case errors.Is(err, domainerrors.ErrInvalidQuantity):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid return or exchange quantity",
			})
		case errors.Is(err, domainerrors.ErrProductNotAvailable):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "requested exchange product is unavailable",
			})
		case errors.Is(err, domainerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid return or exchange request",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to create return or exchange request",
			})
		}
		return
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "return or exchange request created successfully",
		Data:    response.ToReturnRequestResponse(returnRequest),
	})
}
func (h *ReturnRequestHandler) GetByID(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid return request id",
		})
		return
	}
	returnRequest, err := h.returnRequestUseCase.GetByID(
		c.Request.Context(), uint(id), uint(userID),
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "return request not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch return request",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "return request fetched successfully",
		Data:    response.ToReturnRequestResponse(returnRequest),
	})
}
func (h *ReturnRequestHandler) ListByUserID(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	limit, page := parsePagination(c)
	returnRequests, total, err := h.returnRequestUseCase.ListByUserID(
		c.Request.Context(),
		uint(userID),
		limit,
		page,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch return requests",
		})
		return
	}
	data := make([]response.ReturnRequestResponse, 0, len(returnRequests))
	for i := range returnRequests {
		data = append(
			data,
			response.ToReturnRequestResponse(&returnRequests[i]),
		)
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "return requests fetched successfully",
		Data: gin.H{
			"items": data,
			"total": total,
			"limit": limit,
			"page":  page,
		},
	})
}
func (h *ReturnRequestHandler) List(c *gin.Context) {
	limit, page := parsePagination(c)
	var orderID *uuid.UUID
	if rawOrderID := c.Query("order_id"); rawOrderID != "" {
		parsedID, err := uuid.Parse(rawOrderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid order id",
			})
			return
		}
		orderID = &parsedID
	}
	var status *string
	if rawStatus := c.Query("status"); rawStatus != "" {
		status = &rawStatus
	}
	returnRequests, total, err := h.returnRequestUseCase.List(
		c.Request.Context(),
		orderID,
		status,
		limit,
		page,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch return requests",
		})
		return
	}
	data := make([]response.ReturnRequestResponse, 0, len(returnRequests))
	for i := range returnRequests {
		data = append(
			data,
			response.ToReturnRequestResponse(&returnRequests[i]),
		)
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "return requests fetched successfully",
		Data: gin.H{
			"items": data,
			"total": total,
			"limit": limit,
			"page":  page,
		},
	})
}
func (h *ReturnRequestHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid return request id",
		})
		return
	}
	var req request.UpdateReturnRequestStatusDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}
	returnRequest, err := h.returnRequestUseCase.UpdateStatus(
		c.Request.Context(),
		uint(id),
		req.Status,
		req.AdminNote,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "return request not found",
			})

		case errors.Is(err, domainerrors.ErrInvalidOrderTransition):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "invalid return request status transition",
			})

		case errors.Is(err, domainerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid return request status",
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to update return request status",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "return request status updated successfully",
		Data:    response.ToReturnRequestResponse(returnRequest),
	})
}
