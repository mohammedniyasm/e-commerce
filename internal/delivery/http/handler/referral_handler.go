package handler

import (
	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReferralHandler struct {
	referralUseCase interfaces.ReferralUseCase
	log             *slog.Logger
}

func NewReferralHandler(
	referralUseCase interfaces.ReferralUseCase,
	log *slog.Logger,
) *ReferralHandler {
	return &ReferralHandler{
		referralUseCase: referralUseCase,
		log:             log,
	}
}

func (h *ReferralHandler) GetReferralDashboard(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid limit",
		})
		return
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid page",
		})
		return
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	status := c.Query("status")
	var statusFilter *models.ReferralStatus
	if status != "" {
		switch models.ReferralStatus(status) {
		case models.ReferralStatusPending,
			models.ReferralStatusCompleted:
			value := models.ReferralStatus(status)
			statusFilter = &value
		default:
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid referral status",
			})
			return
		}
	}
	rewardStatus := c.Query("reward_status")
	var rewardStatusFilter *models.ReferralRewardStatus
	if rewardStatus != "" {
		switch models.ReferralRewardStatus(rewardStatus) {
		case models.ReferralRewardStatusPending,
			models.ReferralRewardStatusCompleted:
			value := models.ReferralRewardStatus(rewardStatus)
			rewardStatusFilter = &value
		default:
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid reward status",
			})
			return
		}
	}
	referralCode,
		total,
		pending,
		completed,
		totalRewards,
		referrals,
		filteredTotal,
		err := h.referralUseCase.GetReferralDashboard(
		c.Request.Context(),
		uint(userID),
		statusFilter,
		rewardStatusFilter,
		limit,
		offset,
	)
	if err != nil {
		h.log.ErrorContext(
			c.Request.Context(),
			"failed to fetch referral dashboard",
			"user_id", userID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch referral dashboard",
		})
		return
	}
	referralItems := make([]response.ReferralItemDTO, 0, len(referrals))
	for _, referral := range referrals {
		rewardEarned := 0.0
		if referral.ReferrerRewardStatus == models.ReferralRewardStatusCompleted {
			rewardEarned = referral.ReferrerReward
		}
		referralItems = append(referralItems, response.ReferralItemDTO{
			ID:           referral.ID,
			Name:         referral.ReferredUser.Name,
			Status:       string(referral.Status),
			RewardStatus: string(referral.ReferrerRewardStatus),
			RewardEarned: rewardEarned,
			CreatedAt:    referral.CreatedAt,
			CompletedAt:  referral.CompletedAt,
		})
	}
	dashboard := response.ReferralDashboardResponse{
		ReferralCode: referralCode,
		ReferralLink: "/register?ref=" + referralCode,
		Summary: response.ReferralSummaryDTO{
			TotalReferrals:     total,
			PendingReferrals:   pending,
			CompletedReferrals: completed,
			TotalRewardsEarned: totalRewards,
		},
		Referrals: referralItems,
		Pagination: response.PaginationDTO{
			Total:  filteredTotal,
			Limit:  limit,
			Offset: offset,
		},
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "referral dashboard fetched successfully",
		Data:    dashboard,
	})
}
func (h *ReferralHandler) GetAdminReferralDashboard(c *gin.Context) {
	total, pending, completed, referrerRewards, referredRewards,
		totalRewards, conversionRate, err :=
		h.referralUseCase.GetAdminReferralDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch referral dashboard",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "referral dashboard fetched successfully",
		Data: response.AdminReferralDashboardResponse{
			TotalReferrals:      total,
			PendingReferrals:    pending,
			CompletedReferrals:  completed,
			ReferrerRewardsPaid: referrerRewards,
			ReferredRewardsPaid: referredRewards,
			TotalRewardsPaid:    totalRewards,
			ConversionRate:      conversionRate,
		},
	})
}
func (h *ReferralHandler) GetTopReferrers(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid limit",
		})
		return
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	referrers, err := h.referralUseCase.GetTopReferrers(
		c.Request.Context(),
		limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch top referrers",
		})
		return
	}
	result := make([]response.TopReferrerDTO, 0, len(referrers))
	for _, referrer := range referrers {
		result = append(result, response.TopReferrerDTO{
			ReferrerID:          referrer.ReferrerID,
			Name:                referrer.Name,
			SuccessfulReferrals: referrer.SuccessfulReferrals,
			RewardsEarned:       referrer.RewardsEarned,
		})
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "top referrers fetched successfully",
		Data: response.TopReferrersResponse{
			Referrers: result,
		},
	})
}
func (h *ReferralHandler) GetReferrals(c *gin.Context) {
	var status *models.ReferralStatus
	var rewardStatus *models.ReferralRewardStatus
	var referrerID *uint
	var startDate *time.Time
	var endDate *time.Time
	if value := c.Query("status"); value != "" {
		s := models.ReferralStatus(value)
		if s != models.ReferralStatusPending &&
			s != models.ReferralStatusCompleted {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid referral status",
			})
			return
		}
		status = &s
	}
	if value := c.Query("reward_status"); value != "" {
		s := models.ReferralRewardStatus(value)
		if s != models.ReferralRewardStatusPending &&
			s != models.ReferralRewardStatusCompleted {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid reward status",
			})
			return
		}
		rewardStatus = &s
	}
	if value := c.Query("referrer_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 32)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid referrer id",
			})
			return
		}
		parsedID := uint(id)
		referrerID = &parsedID
	}
	if value := c.Query("start_date"); value != "" {
		parsedDate, err := time.Parse("2006-01-02", value)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid start date",
			})
			return
		}
		startDate = &parsedDate
	}
	if value := c.Query("end_date"); value != "" {
		parsedDate, err := time.Parse("2006-01-02", value)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid end date",
			})
			return
		}
		nextDay := parsedDate.AddDate(0, 0, 1)
		endDate = &nextDay
	}
	if startDate != nil && endDate != nil && !startDate.Before(*endDate) {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid date range",
		})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid page",
		})
		return
	}
	if page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid limit",
		})
		return
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	referrals, total, err := h.referralUseCase.GetReferrals(
		c.Request.Context(),
		status,
		rewardStatus,
		referrerID,
		startDate,
		endDate,
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch referrals",
		})
		return
	}
	result := make([]response.AdminReferralItemDTO, 0, len(referrals))
	for _, referral := range referrals {
		result = append(result, response.AdminReferralItemDTO{
			ID:                   referral.ID,
			ReferrerID:           referral.ReferrerID,
			ReferrerName:         referral.Referrer.Name,
			ReferredUserID:       referral.ReferredUserID,
			ReferredUserName:     referral.ReferredUser.Name,
			Status:               string(referral.Status),
			ReferrerReward:       referral.ReferrerReward,
			ReferredReward:       referral.ReferredReward,
			ReferrerRewardStatus: string(referral.ReferrerRewardStatus),
			ReferredRewardStatus: string(referral.ReferredRewardStatus),
			CreatedAt:            referral.CreatedAt,
			CompletedAt:          referral.CompletedAt,
		})
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "referrals fetched successfully",
		Data: response.AdminReferralsResponse{
			Referrals: result,
			Pagination: response.PaginationDTO{
				Total:  total,
				Limit:  limit,
				Offset: offset,
			},
		},
	})
}
func (h *ReferralHandler) GetReferralSettings(c *gin.Context) {
	settings, err := h.referralUseCase.GetReferralSettings(
		c.Request.Context(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch referral settings",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "referral settings fetched successfully",
		Data: response.ReferralSettingsResponse{
			ReferrerReward:     settings.ReferrerReward,
			ReferredReward:     settings.ReferredReward,
			MinimumOrderAmount: settings.MinimumOrderAmount,
			IsActive:           settings.IsActive,
			UpdatedAt:          settings.UpdatedAt,
		},
	})
}
func (h *ReferralHandler) UpdateReferralSettings(c *gin.Context) {
	var req request.UpdateReferralSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}
	settings, err := h.referralUseCase.UpdateReferralSettings(
		c.Request.Context(),
		req.ReferrerReward,
		req.ReferredReward,
		req.MinimumOrderAmount,
		*req.IsActive,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidReferralSettings) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid referral settings",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to update referral settings",
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "referral settings updated successfully",
		Data: response.ReferralSettingsResponse{
			ReferrerReward:     settings.ReferrerReward,
			ReferredReward:     settings.ReferredReward,
			MinimumOrderAmount: settings.MinimumOrderAmount,
			IsActive:           settings.IsActive,
			UpdatedAt:          settings.UpdatedAt,
		},
	})
}
