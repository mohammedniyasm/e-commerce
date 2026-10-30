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
func (h *WishlistHandler) CreateWishlist(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	var req request.CreateWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}
	err := h.wishlistUseCase.CreateWishlist(
		c.Request.Context(),
		uint(userID),
		req.Name,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrWishlistItemAlreadyExists):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "wishlist with this name already exists",
			})

		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to create wishlist",
			})
		}
		return
	}
	c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "wishlist created successfully",
	})
}
func (h *WishlistHandler) GetWishlists(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	wishlists, err := h.wishlistUseCase.GetWishlists(
		c.Request.Context(),
		uint(userID),
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
			Message: "failed to fetch wishlists",
		})
		return
	}
	result := make([]response.WishlistResponse, 0, len(wishlists))
	for i := range wishlists {
		result = append(result, response.ToWishlistResponse(&wishlists[i]))
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlists fetched successfully",
		Data:    result,
	})
}
func (h *WishlistHandler) GetWishlistByID(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	wishlistID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || wishlistID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid wishlist id",
		})
		return
	}
	wishlist, err := h.wishlistUseCase.GetWishlistByID(
		c.Request.Context(),
		uint(userID),
		uint(wishlistID),
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "wishlist not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to fetch wishlist",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist fetched successfully",
		Data:    response.ToWishlistDetailsResponse(wishlist),
	})

}
func (h *WishlistHandler) UpdateWishlist(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	wishlistID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || wishlistID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid wishlist id",
		})
		return
	}
	var req request.UpdateWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}
	err = h.wishlistUseCase.UpdateWishlist(
		c.Request.Context(),
		uint(wishlistID),
		uint(userID),
		req.Name,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "wishlist not found",
			})
		case errors.Is(err, domainerrors.ErrWishlistItemAlreadyExists):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "wishlist with this name already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to update wishlist",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist updated successfully",
	})
}
func (h *WishlistHandler) DeleteWishlist(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	wishlistID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || wishlistID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid wishlist id",
		})
		return
	}
	err = h.wishlistUseCase.DeleteWishlist(
		c.Request.Context(),
		uint(wishlistID),
		uint(userID),
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "wishlist not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to delete wishlist",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist deleted successfully",
	})
}
func (h *WishlistHandler) AddProductToWishlist(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	var req request.AddProductToWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}
	wishlist, err := h.wishlistUseCase.AddProductToWishlist(
		c.Request.Context(),
		req.WishlistName,
		uint(userID),
		req.ProductID,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrInvalidProductID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid product id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "product not found",
			})
		case errors.Is(err, domainerrors.ErrProductNotAvailable):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product is unavailable",
			})
		case errors.Is(err, domainerrors.ErrWishlistItemAlreadyExists):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "product already exists in wishlist",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to add product to wishlist",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "product added to wishlist successfully",
		Data:    response.ToWishlistDetailsResponse(wishlist),
	})
}
func (h *WishlistHandler) RemoveWishlistItem(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	wishlistID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || wishlistID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid wishlist id",
		})
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil || itemID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid wishlist item id",
		})
		return
	}
	wishlist, err := h.wishlistUseCase.RemoveWishlistItem(
		c.Request.Context(),
		uint(wishlistID),
		uint(userID),
		uint(itemID),
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "wishlist or wishlist item not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to remove wishlist item",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist item removed successfully",
		Data:    response.ToWishlistDetailsResponse(wishlist),
	})
}
func (h *WishlistHandler) ClearWishlist(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	wishlistID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || wishlistID == 0 {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid wishlist id",
		})
		return
	}
	wishlist, err := h.wishlistUseCase.ClearWishlist(
		c.Request.Context(),
		uint(wishlistID),
		uint(userID),
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "wishlist not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to clear wishlist",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist cleared successfully",
		Data:    response.ToWishlistDetailsResponse(wishlist),
	})
}
func (h *WishlistHandler) MergeWishlistItems(c *gin.Context) {
	userID, ok := getCartUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Success: false,
			Message: "authentication required",
		})
		return
	}
	var req request.MergeWishlistItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Message: "invalid request",
		})
		return
	}
	wishlist, accepted, skipped, err := h.wishlistUseCase.MergeWishlistItems(
		c.Request.Context(),
		req.WishlistName,
		uint(userID),
		req.ProductIDs,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid user id",
			})
		case errors.Is(err, domainerrors.ErrInvalidProductID):
			c.JSON(http.StatusBadRequest, response.APIResponse{
				Success: false,
				Message: "invalid product id",
			})
		case errors.Is(err, domainerrors.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, response.APIResponse{
				Success: false,
				Message: "one or more wishlist products were not found",
			})
		case errors.Is(err, domainerrors.ErrProductNotAvailable):
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Message: "one or more products are unavailable",
			})
		default:
			c.JSON(http.StatusInternalServerError, response.APIResponse{
				Success: false,
				Message: "failed to merge wishlist",
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "wishlist merged successfully",
		Data: response.MergeWishlistItemsResponse{
			Wishlist: response.ToWishlistDetailsResponse(wishlist),
			Accepted: accepted,
			Skipped:  skipped,
		},
	})
}
