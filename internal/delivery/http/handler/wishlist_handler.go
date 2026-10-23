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

type WishlistHandler struct {
	wishlistUseCase interfaces.WishlistUseCase
}

func NewWishlistHandler(
	wishlistUseCase interfaces.WishlistUseCase,
) *WishlistHandler {
	return &WishlistHandler{
		wishlistUseCase: wishlistUseCase,
	}
}

func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}

	wishlist, err := h.wishlistUseCase.GetWishlist(
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

		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to fetch wishlist",
			// Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist fetched successfully",
		Data:    response.ToWishlistResponse(wishlist),
	})
}
func (h *WishlistHandler) AddProduct(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}

	var req request.AddWishlistProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			// Error:   err.Error(),
		})
		return
	}
	wishlist, err := h.wishlistUseCase.AddProduct(
		c.Request.Context(),
		userID,
		req.ProductID,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidProductID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid product id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrProductNotAvailable) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product is unavailable",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrWishlistItemAlreadyExists) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product already exists in wishlist",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to add product to wishlist",
			// Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "product added to wishlist successfully",
		Data:    response.ToWishlistResponse(wishlist),
	})
}
func (h *WishlistHandler) RemoveItem(c *gin.Context) {
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
			Message: "invalid wishlist item id",
		})
		return
	}
	wishlist, err := h.wishlistUseCase.RemoveItem(
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
				Message: "wishlist or wishlist item not found",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrProductNotAvailable) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product is unavailable",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to remove wishlist item",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist item removed successfully",
		Data:    response.ToWishlistResponse(wishlist),
	})
}
func (h *WishlistHandler) ClearWishlist(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			// Error:   "unauthorized",
		})
		return
	}
	wishlist, err := h.wishlistUseCase.ClearWishlist(
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
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to clear wishlist",
			// Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist cleared successfully",
		Data:    response.ToWishlistResponse(wishlist),
	})
}
func (h *WishlistHandler) MergeGuestWishlist(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
			Error:   "unauthorized",
		})
		return
	}
	var req request.MergeGuestWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
			// Error:   err.Error(),
		})
		return
	}

	wishlist, err := h.wishlistUseCase.MergeGuestWishlist(
		c.Request.Context(),
		userID,
		req.ProductIDs,
	)

	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrInvalidProductID) {
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid product id",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrRecordNotFound) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "one or more wishlist products no longer exist",
			})
			return
		}
		if errors.Is(err, domainerrors.ErrProductNotAvailable) {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product is unavailable",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Success: false,
			Message: "failed to merge guest wishlist",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "guest wishlist merged successfully",
		Data:    response.ToWishlistResponse(wishlist),
	})
}
