package handler

import (
	"errors"
	"io"
	"net/http"

	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentUseCase interfaces.PaymentUseCase
}

func NewPaymentHandler(
	paymentUseCase interfaces.PaymentUseCase,
) *PaymentHandler {
	return &PaymentHandler{
		paymentUseCase: paymentUseCase,
	}
}

func (h *PaymentHandler) HandleRazorpayWebhook(c *gin.Context) {
	signature := c.GetHeader("X-Razorpay-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "missing webhook signature",
		})
		return
	}
	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "failed to read webhook payload",
		})
		return
	}
	err = h.paymentUseCase.HandleRazorpayWebhook(
		c.Request.Context(),
		payload,
		signature,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid webhook request",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "payment not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "webhook processing failed",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "webhook processed successfully",
	})
}
func (h *PaymentHandler) CreateRazorpayOrder(c *gin.Context) {
	var req request.CreatePaymentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	id, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}
	razorpayOrderID, err := h.paymentUseCase.CreateRazorpayOrder(
		c.Request.Context(),
		id,
		req.OrderID,
	)
	if err != nil {
		// Map your existing domain errors here.
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create payment order",
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "payment order created successfully",
		Data: gin.H{
			"razorpay_order_id": razorpayOrderID,
		},
	})
}
func (h *PaymentHandler) VerifyRazorpayPayment(c *gin.Context) {
	var req request.PaymentVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	id, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	err := h.paymentUseCase.VerifyRazorpayPayment(
		c.Request.Context(),
		id,
		req.OrderID,
		req.RazorpayOrderID,
		req.RazorpayPaymentID,
		req.RazorpaySignature,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "payment verification failed",
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "payment verified successfully",
	})
}
