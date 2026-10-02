package handler

import (
	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/delivery/http/middleware"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileUseCase interfaces.ProfileUseCase
}

func NewProfileHandler(profileUseCase interfaces.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{
		profileUseCase: profileUseCase,
	}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userIDValue, exists := c.Get(middleware.UserIdKey)
	if !exists {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "user identity not found",
		})
		return
	}
	userIDString, ok := userIDValue.(string)
	if !ok {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}
	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}
	user, err := h.profileUseCase.GetProfile(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(500, response.APIResponse{
			Success: false,
			Message: "failed to retrieve profile",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "profile retrieved successfully",
		Data: response.ProfileResponse{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			ProfileImage:    user.ProfileImage,
			EmailVerifiedAt: user.EmailVerifiedAt,
		},
	})
}
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var req request.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	userIDValue, exists := c.Get(middleware.UserIdKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "user identity not found",
		})
		return
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	userID64, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.profileUseCase.UpdateProfile(
		c.Request.Context(),
		uint(userID64),
		req.Name,
		req.Phone,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidName):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid name",
				Error:   err.Error(),
			})
		case errors.Is(err, domainerrors.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid phone number",
				Error:   err.Error(),
			})
		case errors.Is(err, domainerrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "user not found",
				Error:   err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to update profile",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "profile updated successfully",
		Data: response.ProfileResponse{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			ProfileImage:    user.ProfileImage,
			EmailVerifiedAt: user.EmailVerifiedAt,
		},
	})
}
func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	var req request.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	userIDValue, exists := c.Get(middleware.UserIdKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "user identity not found",
		})
		return
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	userID64, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	err = h.profileUseCase.ChangePassword(
		c.Request.Context(),
		uint(userID64),
		req.CurrentPassword,
		req.NewPassword,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, response.APIResponse{
				Success: false,
				Message: "current password is incorrect",
			})

		case errors.Is(err, domainerrors.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "new password does not meet password requirements",
				Error:   err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to change password",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "password changed successfully",
	})
}
func (h *ProfileHandler) SendEmailChangeOTP(c *gin.Context) {
	var req request.ChangeEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	userIDValue, exists := c.Get(middleware.UserIdKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "user identity not found",
		})
		return
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	err = h.profileUseCase.SendEmailChangeOTP(
		c.Request.Context(),
		uint(userID),
		req.NewEmail,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "email is already in use",
			})

		case errors.Is(err, domainerrors.ErrUserNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "user not found",
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to send email verification OTP",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "email verification OTP sent successfully",
	})
}
func (h *ProfileHandler) VerifyEmailChangeOTP(c *gin.Context) {
	var req request.VerifyEmailChangeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}

	userIDValue, exists := c.Get(middleware.UserIdKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "user identity not found",
		})
		return
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	err = h.profileUseCase.VerifyEmailChangeOTP(
		c.Request.Context(),
		uint(userID),
		req.OTP,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidOTP):
			c.JSON(http.StatusUnauthorized, response.APIResponse{
				Success: false,
				Message: "invalid or expired OTP",
			})

		case errors.Is(err, domainerrors.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "email is already in use",
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to verify email change",
				Error:   err.Error(),
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "email changed successfully",
	})
}
func (h *ProfileHandler) UploadProfileImage(c *gin.Context) {
	userIDValue, exists := c.Get(middleware.UserIdKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "user identity not found",
		})
		return
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	userID64, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "profile image is required",
			Error:   err.Error(),
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "failed to open profile image",
			Error:   err.Error(),
		})
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")

	imageURL, err := h.profileUseCase.UploadProfileImage(
		c.Request.Context(),
		uint(userID64),
		file,
		fileHeader.Size,
		contentType,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidProfileImage) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid profile image",
				Error:   err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to upload profile image",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "profile image updated successfully",
		Data: map[string]string{
			"profile_image": imageURL,
		},
	})
}
func (h *ProfileHandler) DeleteProfileImage(c *gin.Context) {
	userIDValue, exists := c.Get(middleware.UserIdKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "user identity not found",
		})
		return
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}

	userID64, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil || userID64 == 0 {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}
	err = h.profileUseCase.DeleteProfileImage(
		c.Request.Context(),
		uint(userID64),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "failed to delete profile image",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Profile image deleted successfully",
	})
}
