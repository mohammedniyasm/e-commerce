package handler

import (
	"errors"
	"net/http"
	"strconv"

	"ecommerce/internal/delivery/http/dto/request"
	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/delivery/http/middleware"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AddressHandler struct {
	addressUseCase interfaces.AddressUseCase
}

func NewAddressHandler(
	addressUseCase interfaces.AddressUseCase,
) *AddressHandler {
	return &AddressHandler{
		addressUseCase: addressUseCase,
	}
}

func (h *AddressHandler) AddAddress(c *gin.Context) {
	var req request.AddAddressRequest

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

	address := &models.Address{
		Name:         req.Name,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		PostalCode:   req.PostalCode,
		State:        req.State,
		Country:      req.Country,
		IsDefault:    req.IsDefault,
	}

	createdAddress, err := h.addressUseCase.AddAddress(
		c.Request.Context(),
		uint(userID64),
		address,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidName):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid address name",
				Error:   err.Error(),
			})

		case errors.Is(err, domainerrors.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid address phone",
				Error:   err.Error(),
			})

		case errors.Is(err, domainerrors.ErrInvalidAddress):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid address",
				Error:   err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to create address",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "address created successfully",
		Data: response.AddressResponse{
			ID:           createdAddress.ID,
			Name:         createdAddress.Name,
			Phone:        createdAddress.Phone,
			AddressLine1: createdAddress.AddressLine1,
			AddressLine2: createdAddress.AddressLine2,
			City:         createdAddress.City,
			PostalCode:   createdAddress.PostalCode,
			State:        createdAddress.State,
			Country:      createdAddress.Country,
			IsDefault:    createdAddress.IsDefault,
			CreatedAt:    createdAddress.CreatedAt,
			UpdatedAt:    createdAddress.UpdatedAt,
		},
	})
}
func (h *AddressHandler) GetAddresses(c *gin.Context) {
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

	addresses, err := h.addressUseCase.GetAddresses(
		c.Request.Context(),
		uint(userID64),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to retrieve addresses",
			Error:   err.Error(),
		})
		return
	}

	data := make([]response.AddressResponse, 0, len(addresses))

	for _, address := range addresses {
		data = append(data, response.AddressResponse{
			ID:           address.ID,
			Name:         address.Name,
			Phone:        address.Phone,
			AddressLine1: address.AddressLine1,
			AddressLine2: address.AddressLine2,
			City:         address.City,
			PostalCode:   address.PostalCode,
			State:        address.State,
			Country:      address.Country,
			IsDefault:    address.IsDefault,
			CreatedAt:    address.CreatedAt,
			UpdatedAt:    address.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "addresses retrieved successfully",
		Data:    data,
	})
}
func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	var req request.UpdateAddressRequest

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

	addressID64, err := strconv.ParseUint(
		c.Param("id"),
		10,
		64,
	)
	if err != nil || addressID64 == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid address id",
		})
		return
	}

	address := &models.Address{
		Name:         req.Name,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		PostalCode:   req.PostalCode,
		State:        req.State,
		Country:      req.Country,
		IsDefault:    req.IsDefault,
	}

	updatedAddress, err := h.addressUseCase.UpdateAddress(
		c.Request.Context(),
		uint(userID64),
		uint(addressID64),
		address,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidName):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid address name",
				Error:   err.Error(),
			})

		case errors.Is(err, domainerrors.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid address phone",
				Error:   err.Error(),
			})

		case errors.Is(err, domainerrors.ErrInvalidAddress):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid address",
				Error:   err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to update address",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "address updated successfully",
		Data: response.AddressResponse{
			ID:           updatedAddress.ID,
			Name:         updatedAddress.Name,
			Phone:        updatedAddress.Phone,
			AddressLine1: updatedAddress.AddressLine1,
			AddressLine2: updatedAddress.AddressLine2,
			City:         updatedAddress.City,
			PostalCode:   updatedAddress.PostalCode,
			State:        updatedAddress.State,
			Country:      updatedAddress.Country,
			IsDefault:    updatedAddress.IsDefault,
			CreatedAt:    updatedAddress.CreatedAt,
			UpdatedAt:    updatedAddress.UpdatedAt,
		},
	})
}
func (h *AddressHandler) DeleteAddress(c *gin.Context) {
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

	addressID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || addressID64 == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid address id",
		})
		return
	}

	err = h.addressUseCase.DeleteAddress(
		c.Request.Context(),
		uint(userID64),
		uint(addressID64),
	)
	if err != nil {
		c.JSON(http.StatusNotFound, response.APIResponse{
			Success: false,
			Message: "address not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "address deleted successfully",
	})
}
func (h *AddressHandler) SetDefaultAddress(c *gin.Context) {
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

	addressID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || addressID64 == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid address id",
		})
		return
	}

	err = h.addressUseCase.SetDefaultAddress(
		c.Request.Context(),
		uint(userID64),
		uint(addressID64),
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "address not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to set default address",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "default address updated successfully",
	})
}
