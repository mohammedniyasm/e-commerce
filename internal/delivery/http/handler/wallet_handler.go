package handler

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"

	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/infrastructure/payment/razorpay"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletUseCase interfaces.WalletUseCase
	razorpayKeyID string
	webhookSecret string
}

func NewWalletHandler(
	walletUseCase interfaces.WalletUseCase,
	razorpayKeyID string,
	webhookSecret string,
) *WalletHandler {
	return &WalletHandler{
		walletUseCase: walletUseCase,
		razorpayKeyID: razorpayKeyID,
		webhookSecret: webhookSecret,
	}
}
func (h *WalletHandler) GetWallet(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	wallet, err := h.walletUseCase.GetWallet(
		c.Request.Context(),
		uint(userID),
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidInput),
			errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch wallet",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wallet fetched successfully",
		Data:    response.ToWalletResponse(wallet),
	})
}
func (h *WalletHandler) GetTransactionByID(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || transactionID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid transaction id",
		})
		return
	}
	transaction, err := h.walletUseCase.GetTransactionByID(
		c.Request.Context(),
		uint(userID),
		uint(transactionID),
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid input",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "wallet transaction not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch wallet transaction",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wallet transaction fetched successfully",
		Data:    response.ToWalletTransactionResponse(transaction),
	})
}
func (h *WalletHandler) ListTransactions(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid limit",
		})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid page",
		})
		return
	}
	sort := c.Query("sort")
	var sortValue *string
	if sort != "" {
		sortValue = &sort
	}
	transactions, total, err := h.walletUseCase.ListTransactions(
		c.Request.Context(),
		uint(userID),
		limit,
		page,
		sortValue,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch wallet transactions",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wallet transactions fetched successfully",
		Data: gin.H{
			"transactions": response.ToWalletTransactionListResponse(transactions),
			"total":        total,
			"page":         page,
			"limit":        limit,
		},
	})
}
func (h *WalletHandler) GetUserWallet(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid user id",
		})
		return
	}
	wallet, err := h.walletUseCase.GetWallet(
		c.Request.Context(),
		uint(userID),
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "user not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch user wallet",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "user wallet fetched successfully",
		Data:    response.ToWalletResponse(wallet),
	})
}
func (h *WalletHandler) GetUserWalletTransactions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid user id",
		})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid limit",
		})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid page",
		})
		return
	}
	sort := c.Query("sort")
	var sortValue *string
	if sort != "" {
		sortValue = &sort
	}
	transactions, total, err := h.walletUseCase.ListTransactions(
		c.Request.Context(),
		uint(userID),
		limit,
		page,
		sortValue,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch user wallet transactions",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "user wallet transactions fetched successfully",
		Data: gin.H{
			"transactions": response.ToWalletTransactionListResponse(transactions),
			"total":        total,
			"page":         page,
			"limit":        limit,
		},
	})
}
func (h *WalletHandler) CreditUserWallet(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid user id",
		})
		return
	}
	var req request.AdminWalletCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}
	reason := req.Reason
	err = h.walletUseCase.CreditWallet(
		c.Request.Context(),
		uint(userID),
		req.Amount,
		models.WalletSourceAdminAdjustment,
		nil,
		&reason,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid credit details",
			})
		case errors.Is(err, domainerrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "user not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to credit user wallet",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "user wallet credited successfully",
	})
}
func (h *WalletHandler) CreateTopUpOrder(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	var req request.CreateTopUpOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid wallet top-up request",
		})
		return
	}
	gatewayOrderID, err := h.walletUseCase.CreateTopUpOrder(
		c.Request.Context(),
		uint(userID),
		req.Amount,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid top-up amount",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create wallet top-up order",
		})
		return
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "wallet top-up order created successfully",
		Data: gin.H{
			"razorpay_key_id":  h.razorpayKeyID,
			"gateway_order_id": gatewayOrderID,
			"amount":           int64(math.Round(req.Amount * 100)),
			"currency":         "INR",
		},
	})
}
func (h *WalletHandler) VerifyTopUpPayment(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	var req request.VerifyTopUpPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid payment verification request",
		})
		return
	}
	err := h.walletUseCase.VerifyTopUpPayment(
		c.Request.Context(),
		uint(userID),
		req.GatewayOrderID,
		req.GatewayPaymentID,
		req.Signature,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid payment details",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "wallet top-up not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to verify wallet top-up payment",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wallet topped up successfully",
	})
}
func (h *WalletHandler) RazorpayWebhook(c *gin.Context) {
	const maxBodySize = 1 << 20 // 1 MB
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxBodySize,
	)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid webhook payload",
		})
		return
	}
	signature := c.GetHeader("X-Razorpay-Signature")
	if !razorpay.VerifyWebhookSignature(
		body,
		signature,
		h.webhookSecret,
	) {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid webhook signature",
		})
		return
	}
	var webhook models.RazorpayWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid webhook event",
		})
		return
	}
	if webhook.Event != "payment.captured" {
		c.JSON(http.StatusOK, response.APIResponse{
			Success: true,
			Message: "webhook event ignored",
		})
		return
	}
	payment := webhook.Payload.Payment.Entity
	if payment.ID == "" || payment.OrderID == "" {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "missing payment details",
		})
		return
	}
	err = h.walletUseCase.ProcessTopUpWebhook(
		c.Request.Context(),
		payment.OrderID,
		payment.ID,
		payment.Amount,
		payment.Currency,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "webhook processing failed",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "webhook processed successfully",
	})
}
