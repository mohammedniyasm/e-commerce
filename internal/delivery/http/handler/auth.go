package handler

import (
	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase interfaces.AuthUseCase
}

func NewAuthHandler(authUsecase interfaces.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
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
func (h *AuthHandler) Login(c *gin.Context){
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req);err!=nil{
		c.JSON(400,response.APIResponse{
			Success: false,
			Message: "invalid request",
			Error: err.Error(),
		})
		return
	}
	accessToken,refreshToken,err:=h.authUsecase.Login(c.Request.Context(),req.Email,req.Password)
	if err != nil{
		c.JSON(401,response.APIResponse{
			Success: false,
			Message: "login failed",
			Error: err.Error(),
		})
		return
	}
	c.JSON(200,response.APIResponse{
		Success: true,
		Message: "login successful",
		Data: response.LoginResponse{
			AccessToken: accessToken,
			RefreshToken: refreshToken,
		},
	})
}