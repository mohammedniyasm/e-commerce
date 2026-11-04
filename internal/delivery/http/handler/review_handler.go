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

type ReviewHandler struct {
	reviewUseCase interfaces.ReviewUseCase
}

func NewReviewHandler(
	reviewUseCase interfaces.ReviewUseCase,
) *ReviewHandler {
	return &ReviewHandler{
		reviewUseCase: reviewUseCase,
	}
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		unauthorized(c)
		return
	}
	var req request.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid review request")
		return
	}
	review, err := h.reviewUseCase.CreateReview(
		c.Request.Context(),
		uint(userID),
		req.OrderItemID,
		req.Rating,
		req.Comment,
	)
	if err != nil {
		h.handleReviewError(c, err, "failed to create review")
		return
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "review created successfully",
		Data:    response.ToReviewResponse(review),
	})
}
func (h *ReviewHandler) GetMyReview(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		unauthorized(c)
		return
	}
	reviewID, ok := h.parseReviewID(c)
	if !ok {
		return
	}
	review, err := h.reviewUseCase.GetMyReview(
		c.Request.Context(),
		uint(userID),
		reviewID,
	)
	if err != nil {
		h.handleReviewError(c, err, "failed to fetch review")
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "review fetched successfully",
		Data:    response.ToReviewResponse(review),
	})
}
func (h *ReviewHandler) ListMyReviews(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		unauthorized(c)
		return
	}
	limit, page := parsePagination(c)
	reviews, total, err := h.reviewUseCase.ListMyReviews(
		c.Request.Context(),
		uint(userID),
		page,
		limit,
	)
	if err != nil {
		h.handleReviewError(c, err, "failed to fetch your reviews")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "reviews fetched successfully",
		Data: response.ReviewListResponse{
			Reviews:    response.ToReviewResponses(reviews),
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: calculateTotalPages(total, limit),
		},
	})
}
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		unauthorized(c)
		return
	}
	reviewID, ok := h.parseReviewID(c)
	if !ok {
		return
	}
	var req request.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid review request")
		return
	}
	err := h.reviewUseCase.UpdateReview(
		c.Request.Context(),
		uint(userID),
		reviewID,
		req.Rating,
		req.Comment,
	)
	if err != nil {
		h.handleReviewError(c, err, "failed to update review")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "review updated successfully",
	})
}
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		unauthorized(c)
		return
	}
	reviewID, ok := h.parseReviewID(c)
	if !ok {
		return
	}
	err := h.reviewUseCase.DeleteReview(
		c.Request.Context(),
		uint(userID),
		reviewID,
	)
	if err != nil {
		h.handleReviewError(c, err, "failed to delete review")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "review deleted successfully",
	})
}
func (h *ReviewHandler) ListProductReviews(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || productID == 0 {
		badRequest(c, "invalid product id")
		return
	}
	limit, page := parsePagination(c)
	reviews, total, err := h.reviewUseCase.ListProductReviews(
		c.Request.Context(),
		uint(productID),
		page,
		limit,
	)
	if err != nil {
		h.handleReviewError(c, err, "failed to fetch product reviews")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product reviews fetched successfully",
		Data: response.ReviewListResponse{
			Reviews:    response.ToReviewResponses(reviews),
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: calculateTotalPages(total, limit),
		},
	})
}
func (h *ReviewHandler) ListAllReviews(c *gin.Context) {
	limit, page := parsePagination(c)
	reviews, total, err := h.reviewUseCase.ListAllReviews(
		c.Request.Context(),
		page,
		limit,
	)
	if err != nil {
		h.handleReviewError(c, err, "failed to fetch reviews")
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "reviews fetched successfully",
		Data: response.ReviewListResponse{
			Reviews:    response.ToReviewResponses(reviews),
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: calculateTotalPages(total, limit),
		},
	})
}
func (h *ReviewHandler) SetReviewVisibility(c *gin.Context) {
	reviewID, ok := h.parseReviewID(c)
	if !ok {
		return
	}
	var req request.UpdateReviewVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "is_visible is required")
		return
	}
	err := h.reviewUseCase.SetReviewVisibility(
		c.Request.Context(),
		reviewID,
		*req.IsVisible,
	)
	if err != nil {
		h.handleReviewError(c, err, "failed to update review visibility")
		return
	}
	message := "review hidden successfully"
	if *req.IsVisible {
		message = "review restored successfully"
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: message,
	})
}
func (h *ReviewHandler) parseReviewID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		badRequest(c, "invalid review id")
		return 0, false
	}
	return uint(id), true
}
func (h *ReviewHandler) handleReviewError(c *gin.Context, err error, defaultMessage string) {
	switch {
	case errors.Is(err, domainerrors.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "review or related record not found",
		})
	case errors.Is(err, domainerrors.ErrDuplicateRecord):
		c.JSON(http.StatusConflict, response.APIResponse{
			Success: false,
			Message: "you have already reviewed this purchased item",
		})
	case errors.Is(err, domainerrors.ErrInvalidInput):
		badRequest(c, "invalid review details or product is not eligible for review")
	default:
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: defaultMessage,
		})
	}
}
func unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, response.APIResponse{
		Success: false,
		Message: "authentication required",
	})
}
func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, response.APIResponse{
		Success: false,
		Message: message,
	})
}
func calculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return int((total + int64(limit) - 1) / int64(limit))
}
