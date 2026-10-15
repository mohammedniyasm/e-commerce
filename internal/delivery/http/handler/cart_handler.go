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

type CartHandler struct {
	cartUseCase interfaces.CartUseCase
}

func NewCartHandler(
	cartUseCase interfaces.CartUseCase,
) *CartHandler {
	return &CartHandler{
		cartUseCase: cartUseCase,
	}
}
func getCartUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch id := value.(type) {
	case int64:
		return id, id > 0

	case int:
		return int64(id), id > 0

	case uint:
		if id == 0 {
			return 0, false
		}
		return int64(id), true

	case string:
		parsedID, err := strconv.ParseInt(id, 10, 64)
		if err != nil || parsedID <= 0 {
			return 0, false
		}
		return parsedID, true

	default:
		return 0, false
	}
}
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}
	cart, err := h.cartUseCase.GetCart(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "cart not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch cart",
			// Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "cart fetched successfully",
		Data:    response.ToCartResponse(cart),
	})
}
func (h *CartHandler) AddItem(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}
	var req request.AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			// Error:   err.Error(),
		})
		return
	}
	cart, err := h.cartUseCase.AddItem(
		c.Request.Context(),
		userID,
		req.VariantID,
		req.Quantity,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidVariantID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid variant id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidQuantity) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "quantity must be greater than zero",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product variant not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrProductNotAvailable) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product or variant is unavailable",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInsufficientStock) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "insufficient stock",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrCartItemAlreadyExists) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "variant already exists in cart",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to add item to cart",
			// Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "item added to cart successfully",
		Data:    response.ToCartResponse(cart),
	})
}
func (h *CartHandler) UpdateQuantity(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid cart item id",
		})
		return
	}
	var req request.UpdateCartItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			// Error:   err.Error(),
		})
		return
	}
	cart, err := h.cartUseCase.UpdateQuantity(
		c.Request.Context(),
		userID,
		itemID,
		req.Quantity,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidQuantity) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "quantity must be greater than zero",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "cart or cart item not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrProductNotAvailable) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product or variant is unavailable",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInsufficientStock) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "requested quantity exceeds available stock",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to update cart item quantity",
			// Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "cart item quantity updated successfully",
		Data:    response.ToCartResponse(cart),
	})
}
func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}

	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid cart item id",
		})
		return
	}

	cart, err := h.cartUseCase.RemoveItem(
		c.Request.Context(),
		userID,
		itemID,
	)

	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "cart or cart item not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to remove cart item",
			// Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "cart item removed successfully",
		Data:    response.ToCartResponse(cart),
	})
}
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}

	cart, err := h.cartUseCase.ClearCart(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}

		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "cart not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to clear cart",
			// Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "cart cleared successfully",
		Data:    response.ToCartResponse(cart),
	})
}
func (h *CartHandler) ValidateGuestCart(c *gin.Context) {
	var req request.ValidateGuestCartRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid guest cart request",
			// Error:   err.Error(),
		})
		return
	}
	items := make([]interfaces.GuestCartItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, interfaces.GuestCartItemInput{
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		})
	}
	results, err := h.cartUseCase.ValidateGuestCart(
		c.Request.Context(),
		items,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidVariantID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid variant id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidQuantity) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "quantity must be greater than zero",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to validate guest cart",
			// Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "guest cart validated successfully",
		Data:    response.ToGuestCartValidationResponses(results),
	})
}
func (h *CartHandler) MergeCart(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}
	var req request.MergeCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			// Error:   err.Error(),
		})
		return
	}
	items := make([]interfaces.GuestCartItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, interfaces.GuestCartItemInput{
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		})
	}
	cart, err := h.cartUseCase.MergeGuestCart(
		c.Request.Context(),
		userID,
		items,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidVariantID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid variant id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidQuantity) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid quantity",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "one or more guest cart items no longer exist",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrProductNotAvailable) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "one or more guest cart items are unavailable",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInsufficientStock) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "insufficient stock for one or more items",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to merge guest cart",
			// Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "guest cart merged successfully",
		Data:    response.ToCartResponse(cart),
	})
}
