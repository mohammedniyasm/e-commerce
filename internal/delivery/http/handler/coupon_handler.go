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
	"ecommerce/internal/delivery/http/middleware"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CouponHandler struct {
	couponUseCase interfaces.CouponUseCase
	log           *slog.Logger
}

func NewCouponHandler(
	couponUseCase interfaces.CouponUseCase,
	log *slog.Logger,
) *CouponHandler {
	return &CouponHandler{
		couponUseCase: couponUseCase,
		log:           log,
	}
}
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req request.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid coupon request",
			// Error:   err.Error(),
		})
		return
	}
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid start_at; use RFC3339 format",
			// Error:   err.Error(),
		})
		return
	}
	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid expires_at; use RFC3339 format",
			// Error:   err.Error(),
		})
		return
	}
	coupon := &models.Coupon{
		Code:               strings.ToUpper(strings.TrimSpace(req.Code)),
		Description:        req.Description,
		DiscountType:       strings.ToLower(strings.TrimSpace(req.DiscountType)),
		DiscountValue:      req.DiscountValue,
		MinimumOrderAmount: req.MinimumOrderAmount,
		MaximumDiscount:    req.MaximumDiscount,
		UsageLimit:         req.UsageLimit,
		UsageLimitPerUser:  req.UsageLimitPerUser,
		StartAt:            startAt,
		ExpiresAt:          expiresAt,
		IsActive:           req.IsActive,
	}
	created, err := h.couponUseCase.CreateCoupon(c.Request.Context(), coupon)
	if err != nil {
		h.handleCouponError(c, err)
		return
	}
	h.log.Info("coupon created", "coupon_id", created.ID, "code", created.Code)
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "Coupon created successfully",
		Data:    response.ToCouponResponse(created),
	})
}
func (h *CouponHandler) GetCoupon(c *gin.Context) {
	couponID, ok := h.parseCouponID(c)
	if !ok {
		return
	}
	coupon, err := h.couponUseCase.GetCoupon(c.Request.Context(), couponID)
	if err != nil {
		h.handleCouponError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Coupon retrieved successfully",
		Data:    response.ToCouponResponse(coupon),
	})
}
func (h *CouponHandler) UpdateCoupon(c *gin.Context) {
	couponID, ok := h.parseCouponID(c)
	if !ok {
		return
	}
	var req request.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid coupon request",
			Error:   err.Error(),
		})
		return
	}
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid start_at; use RFC3339 format",
			// Error:   err.Error(),
		})
		return
	}
	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid expires_at; use RFC3339 format",
			// Error:   err.Error(),
		})
		return
	}
	coupon := &models.Coupon{
		ID:                 couponID,
		Code:               strings.ToUpper(strings.TrimSpace(req.Code)),
		Description:        req.Description,
		DiscountType:       strings.ToLower(strings.TrimSpace(req.DiscountType)),
		DiscountValue:      req.DiscountValue,
		MinimumOrderAmount: req.MinimumOrderAmount,
		MaximumDiscount:    req.MaximumDiscount,
		UsageLimit:         req.UsageLimit,
		UsageLimitPerUser:  req.UsageLimitPerUser,
		StartAt:            startAt,
		ExpiresAt:          expiresAt,
		IsActive:           req.IsActive,
	}
	updated, err := h.couponUseCase.UpdateCoupon(c.Request.Context(), coupon)
	if err != nil {
		h.handleCouponError(c, err)
		return
	}
	h.log.Info("coupon updated", "coupon_id", updated.ID)
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Coupon updated successfully",
		Data:    response.ToCouponResponse(updated),
	})
}
func (h *CouponHandler) DeleteCoupon(c *gin.Context) {
	couponID, ok := h.parseCouponID(c)
	if !ok {
		return
	}
	if err := h.couponUseCase.DeleteCoupon(c.Request.Context(), couponID); err != nil {
		h.handleCouponError(c, err)
		return
	}
	h.log.Info("coupon deleted", "coupon_id", couponID)
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Coupon deleted successfully",
	})
}
func (h *CouponHandler) ToggleCouponActive(c *gin.Context) {
	couponID, ok := h.parseCouponID(c)
	if !ok {
		return
	}
	if err := h.couponUseCase.ToggleCouponActive(c.Request.Context(), couponID); err != nil {
		h.handleCouponError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Coupon active status toggled successfully",
	})
}
func (h *CouponHandler) ListCoupons(c *gin.Context) {
	var req request.ListCouponRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	var activeOnly *bool
	if c.Query("active_only") != "" {
		active := false
		switch strings.ToLower(c.Query("active_only")) {
		case "true":
			active = true
		case "false":
			active = false
		default:
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "active_only must be true or false",
			})
			return
		}
		activeOnly = &active
	}
	var sort *string
	if req.Sort != "" {
		allowedSorts := map[string]bool{
			"new": true,
			"old": true,
			// "start_at":   true,
			// "expires_at": true,
			// "code":       true,
		}
		if !allowedSorts[req.Sort] {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "Unsupported sort field",
			})
			return
		}
		sortValue := req.Sort
		sort = &sortValue
	}
	coupons, total, err := h.couponUseCase.ListCoupons(
		c.Request.Context(),
		strings.TrimSpace(req.Search),
		activeOnly,
		req.Limit,
		req.Page,
		sort,
	)
	if err != nil {
		h.handleCouponError(c, err)
		return
	}
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Coupons retrieved successfully",
		Data: response.CouponListResponse{
			Coupons:    response.ToCouponResponses(coupons),
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	var req request.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid coupon validation request",
		})
		return
	}
	if req.OrderAmount <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "order_amount must be greater than zero",
		})
		return
	}
	coupon, discount, err := h.couponUseCase.ValidateCoupon(
		c.Request.Context(),
		strings.ToUpper(strings.TrimSpace(req.Code)),
		userID,
		req.OrderAmount,
	)
	if err != nil {
		h.handleCouponError(c, err)
		return
	}
	finalAmount := req.OrderAmount - discount
	if finalAmount < 0 {
		finalAmount = 0
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Coupon is valid",
		Data: response.ValidateCouponResponse{
			CouponID:       coupon.ID,
			Code:           coupon.Code,
			DiscountType:   coupon.DiscountType,
			DiscountValue:  coupon.DiscountValue,
			OrderAmount:    req.OrderAmount,
			DiscountAmount: discount,
			FinalAmount:    finalAmount,
		},
	})
}
func (h *CouponHandler) ListCouponUsagesByCouponID(c *gin.Context) {
	couponID, ok := h.parseCouponID(c)
	if !ok {
		return
	}
	limit, page := parsePagination(c)
	usages, total, err := h.couponUseCase.ListCouponUsagesByCouponID(
		c.Request.Context(), couponID, limit, page,
	)
	if err != nil {
		h.handleCouponError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Coupon usages retrieved successfully",
		Data: response.CouponUsageListResponse{
			Usages:     response.ToCouponUsageResponses(usages),
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int((total + int64(limit) - 1) / int64(limit)),
		},
	})
}
func (h *CouponHandler) ListUserCouponUsages(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	limit, page := parsePagination(c)
	usages, total, err := h.couponUseCase.ListCouponUsagesByUserID(
		c.Request.Context(), userID, limit, page,
	)
	if err != nil {
		h.handleCouponError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "User coupon usages retrieved successfully",
		Data: response.CouponUsageListResponse{
			Usages:     response.ToCouponUsageResponses(usages),
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int((total + int64(limit) - 1) / int64(limit)),
		},
	})
}
func (h *CouponHandler) parseCouponID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "Invalid coupon ID",
			// Error:   "Coupon ID must be a valid UUID",
		})
		return uuid.Nil, false
	}
	return id, true
}
func (h *CouponHandler) authenticatedUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get(middleware.UserIdKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return 0, false
	}
	userIDString, ok := value.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Invalid authenticated user ID",
		})
		return 0, false
	}
	userID64, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "Invalid authenticated user ID",
		})
		return 0, false
	}
	return uint(userID64), true
}
func parsePagination(c *gin.Context) (limit, page int) {
	limit, _ = strconv.Atoi(c.Query("limit"))
	page, _ = strconv.Atoi(c.Query("page"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	return limit, page
}
func (h *CouponHandler) handleCouponError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "Coupon or usage record not found",
		})

	case errors.Is(err, domainerrors.ErrCouponAlreadyExists):
		c.JSON(http.StatusConflict, response.APIResponse{
			Success: false,
			Message: "A coupon with this code already exists",
		})

	case errors.Is(err, domainerrors.ErrCouponNotActive),
		errors.Is(err, domainerrors.ErrCouponNotStarted),
		errors.Is(err, domainerrors.ErrCouponExpired),
		errors.Is(err, domainerrors.ErrCouponMinimumOrderAmount),
		errors.Is(err, domainerrors.ErrCouponUsageLimitReached),
		errors.Is(err, domainerrors.ErrInvalidCoupon),
		errors.Is(err, domainerrors.ErrInvalidCouponCode),
		errors.Is(err, domainerrors.ErrCouponDiscountInvalid):
		c.JSON(http.StatusUnprocessableEntity, response.APIResponse{
			Success: false,
			Message: err.Error(),
		})

	case errors.Is(err, domainerrors.ErrInvalidInput),
		errors.Is(err, domainerrors.ErrInvalidCouponID):
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: err.Error(),
		})

	default:
		h.log.Error("coupon request failed", "error", err)
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "An internal server error occurred",
		})
	}
}
