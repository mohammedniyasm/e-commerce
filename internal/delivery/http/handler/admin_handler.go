package handler

import (
	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminUseCase interfaces.AdminUseCase
}

func NewAdminHandler(
	adminUseCase interfaces.AdminUseCase,
) *AdminHandler {
	return &AdminHandler{
		adminUseCase: adminUseCase,
	}
}
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req request.AdminUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid query parameters",
			Error:   err.Error(),
		})
		return
	}
	users, total, err := h.adminUseCase.ListUsers(c.Request.Context(), req.Search, req.Page, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch users",
			Error:   err.Error(),
		})
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}

	if req.Limit < 1 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(req.Limit) - 1) / int64(req.Limit))
	}
	userResponses := make([]response.AdminUserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, response.AdminUserResponse{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			ProfileImage:    user.ProfileImage,
			EmailVerifiedAt: user.EmailVerifiedAt,
			Role:            string(user.Role),
			IsBlocked:       user.IsBlocked,
			LastSeen:        user.LastSeen,
			CreatedAt:       user.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "users fetched successfully",
		Data: response.AdminUserListResponse{
			Users:      userResponses,
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
func (h *AdminHandler) BlockUser(c *gin.Context) {
	userID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid user id",
		})
		return
	}
	err = h.adminUseCase.BlockUser(
		c.Request.Context(),
		uint(userID64),
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to block user",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "user blocked successfully",
	})
}
func (h *AdminHandler) UnblockUser(c *gin.Context) {
	userID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid user id",
		})
		return
	}
	err = h.adminUseCase.UnblockUser(
		c.Request.Context(),
		uint(userID64),
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to unblock user",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "user unblocked successfully",
	})
}
func (h *AdminHandler) AddCustomer(c *gin.Context) {

	var req request.AdminAddCustomerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.adminUseCase.AddCustomer(
		c.Request.Context(),
		req.Name,
		req.Email,
		req.Phone,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to create customer",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "customer created successfully",
		Data: response.AdminUserResponse{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			ProfileImage:    user.ProfileImage,
			Role:            string(user.Role),
			IsBlocked:       user.IsBlocked,
			LastSeen:        user.LastSeen,
			CreatedAt:       user.CreatedAt,
			EmailVerifiedAt: user.EmailVerifiedAt,
		},
	})
}
func (h *AdminHandler) GetCustomer(c *gin.Context) {
	userID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid customer id",
		})
		return
	}
	user, err := h.adminUseCase.GetCustomer(
		c.Request.Context(),
		uint(userID64),
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "customer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch customer",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "customer fetched successfully",
		Data: response.AdminUserResponse{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			ProfileImage:    user.ProfileImage,
			EmailVerifiedAt: user.EmailVerifiedAt,
			Role:            string(user.Role),
			IsBlocked:       user.IsBlocked,
			LastSeen:        user.LastSeen,
			CreatedAt:       user.CreatedAt,
		},
	})
}
func (h *AdminHandler) UpdateCustomer(c *gin.Context) {
	userID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid customer id",
		})
		return
	}
	var req request.AdminUpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.adminUseCase.UpdateCustomer(
		c.Request.Context(),
		uint(userID64),
		req.Name,
		req.Email,
		req.Phone,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "customer not found",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to update customer",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "customer updated successfully",
		Data: response.AdminUserResponse{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			ProfileImage:    user.ProfileImage,
			EmailVerifiedAt: user.EmailVerifiedAt,
			Role:            string(user.Role),
			IsBlocked:       user.IsBlocked,
			LastSeen:        user.LastSeen,
			CreatedAt:       user.CreatedAt,
		},
	})
}
func (h *AdminHandler) DeleteCustomer(c *gin.Context) {

	userID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid customer id",
		})
		return
	}

	err = h.adminUseCase.DeleteCustomer(
		c.Request.Context(),
		uint(userID64),
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "customer not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to delete customer",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "customer deleted successfully",
	})
}
