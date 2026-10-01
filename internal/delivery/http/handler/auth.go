package handler

import (
	"ecommerce/config"
	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/delivery/http/middleware"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase interfaces.AuthUseCase
	cookie      config.CookieConfig
}

func NewAuthHandler(authUsecase interfaces.AuthUseCase, cookie config.CookieConfig) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		cookie:      cookie,
	}
}
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request",
		})
		return
	}
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: &req.Password,
	}
	createdUser, err := h.authUsecase.Register(c, user)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidName) {
			c.JSON(400, response.APIResponse{
				Success: false,
				Message: "Invalid Name",
				Error:   "INVALID_NAME",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidPhone) {
			c.JSON(400, response.APIResponse{
				Success: false,
				Message: "Invalid Phone",
				Error:   "INVALID_PHONE",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrWeakPassword) {
			c.JSON(400, response.APIResponse{
				Success: false,
				Message: "Password does not meet the required strength",
				Error:   "WEAK_PASSWORD",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrEmailAlreadyExists) {
			c.JSON(409, response.APIResponse{
				Success: false,
				Message: "Email already exists",
				Error:   "EMAIL_ALREADY_EXISTS",
			})
			return
		}
		c.JSON(500, response.APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   "INTERNAL_SERVER_ERROR",
		})
		return
	}
	c.JSON(201, response.APIResponse{
		Success: true,
		Message: "user created succefully",
		Data: response.UserReponse{
			ID:    createdUser.ID,
			Name:  createdUser.Name,
			Email: createdUser.Email,
			Phone: createdUser.Phone,
		},
	})
}
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	accessToken, refreshToken, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "login failed",
			Error:   err.Error(),
		})
		return
	}
	SetRefreshTokenCookie(c.Writer, h.cookie, refreshToken, h.cookie.MaxAge)
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "login successful",
		Data: response.LoginResponse{
			AccessToken: accessToken,
		},
	})
}
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(h.cookie.RefreshTokenName)
	if err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "refresh token cookie required",
			Error:   err.Error(),
		})
		return
	}

	accessToken, err := h.authUsecase.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "refresh token failed",
			Error:   err.Error(),
		})
		return
	}
	SetRefreshTokenCookie(c.Writer, h.cookie, refreshToken, h.cookie.MaxAge)
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "Access token refreshed Succefully",
		Data: response.LoginResponse{
			AccessToken: accessToken,
		},
	})
}
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie(h.cookie.RefreshTokenName)
	if err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "refresh token cookie required",
			Error:   err.Error(),
		})
		return
	}
	accessClaimValue, exists := c.Get(middleware.AccessClaimsKey)
	if !exists {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "access claims missing",
		})
		return
	}
	accessClaims, ok := accessClaimValue.(*interfaces.AccessClaims)
	if !ok {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "invalid access claims",
		})
	}
	err = h.authUsecase.Logout(c.Request.Context(), refreshToken, accessClaims)
	if err != nil {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "logout failed",
			Error:   err.Error(),
		})
		return
	}
	clearRefreshTokenCookie(c.Writer, h.cookie)
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "logout successful",
	})
}
func (h *AuthHandler) SendVerificationOTP(c *gin.Context) {
	var req request.SendVerificationOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   domainerrors.ErrInvalidCredentials.Error(),
		})
		return
	}
	err := h.authUsecase.SendVerficationOTP(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrEmailAlreadyVerified) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "email is already verified",
			})
			return
		}
		c.JSON(500, response.APIResponse{
			Success: false,
			Message: "failed to send verification OTP",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "verification OTP send successfully",
	})
}
func (h *AuthHandler) ResendVerificationOTP(c *gin.Context) {
	var req request.SendVerificationOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   domainerrors.ErrInvalidCredentials.Error(),
		})
		return
	}
	err := h.authUsecase.ResendVerificationOTP(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrEmailAlreadyVerified) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "email is already verified",
			})
			return
		}
		c.JSON(500, response.APIResponse{
			Success: false,
			Message: "failed to resend verification OTP",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "verification OTP resent successfully",
	})
}
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req request.VerifyEmailRequest
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
		c.JSON(401, response.APIResponse{
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
	if err != nil {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "invalid user identity",
		})
		return
	}
	if err := h.authUsecase.VerifyEmail(c.Request.Context(), uint(userID), req.OTP); err != nil {
		if errors.Is(err, domainerrors.ErrEmailAlreadyVerified) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "email is already verified",
			})
			return
		}
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "email verification failed",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "email verified successfully",
	})

}
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	err := h.authUsecase.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(500, response.APIResponse{
			Success: false,
			Message: "failed to process password reset request",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "if the account exists, a password reset OTP has been sent",
	})
}
func (h *AuthHandler) VerifyForgotPassword(c *gin.Context) {
	var req request.VerifyForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	resetToken, err := h.authUsecase.VerifyForgotPasswordOTP(c.Request.Context(), req.Email, req.OTP)
	if err != nil {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "invalid or expired OTP",
			Error:   err.Error(),
		})
		return
	}
	setPasswordResetCookie(c.Writer, h.cookie, resetToken, int((10 * time.Minute).Seconds()))
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "OTP verified successfully",
	})
}
func (h *AuthHandler) ResendForgotPasswordOTP(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	err := h.authUsecase.ResendForgotPasswordOTP(
		c.Request.Context(),
		req.Email,
	)
	if err != nil {
		c.JSON(500, response.APIResponse{
			Success: false,
			Message: "failed to process password reset request",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "if the account exists, a password reset OTP has been sent",
	})
}
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	resetToken, err := c.Cookie(h.cookie.PasswordResetTokenName)
	if err != nil {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "password reset session required",
			Error:   err.Error(),
		})
		return
	}
	err = h.authUsecase.ResetPassword(
		c.Request.Context(),
		resetToken,
		req.NewPassword,
	)
	if err != nil {
		c.JSON(401, response.APIResponse{
			Success: false,
			Message: "password reset failed",
			Error:   err.Error(),
		})
		return
	}
	clearPasswordResetCookie(
		c.Writer,
		h.cookie,
	)
	c.JSON(200, response.APIResponse{
		Success: true,
		Message: "password reset successfully",
	})
}
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req request.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error:   err.Error(),
		})
		return
	}
	accessToken, refreshToken, err := h.authUsecase.GoogleLogin(c.Request.Context(),req.IDToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserBlocked) {
			c.JSON(http.StatusForbidden, response.APIResponse{
				Success: false,
				Message: "user account is blocked",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "google authentication failed",
			Error:   err.Error(),
		})
		return
	}
	SetRefreshTokenCookie(c.Writer,h.cookie,refreshToken,h.cookie.MaxAge)
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "google login successful",
		Data: map[string]string{
			"access_token": accessToken,
		},
	})
}
