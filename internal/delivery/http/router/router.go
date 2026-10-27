package router

import (
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/middleware"
	"ecommerce/internal/usecase/interfaces"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func SetupRouter(log *slog.Logger,
	authHandler *handler.AuthHandler,
	jwtServices interfaces.JWTService,
	userRepo interfaces.UserRepository,
	blacklist interfaces.AccessTokenBlacklist,
	rateLimiter interfaces.RateLimiter,
	profileHandler *handler.ProfileHandler,
	addressHandler *handler.AddressHandler,
	adminHandler *handler.AdminHandler,
	categoryHandler *handler.CategoryHandler,
	brandHandler *handler.BrandHandler,
	productHandler *handler.ProductHandler,
	productVariantHandler *handler.ProductVariantHandler,
	productImageHandler *handler.ProductImageHandler,
	storeProductHandler *handler.StoreProductHandler,
	cartHandler *handler.CartHandler,
	wishlistHandler *handler.WishlistHandler,
	notificationHandler *handler.NotificationHandler,
	couponHandler *handler.CouponHandler,
	orderHandler *handler.OrderHandler,
	shipmentHandler *handler.ShipmentHandler,
	returnRequestHandler *handler.ReturnRequestHandler,
) *gin.Engine {
	r := gin.Default()
	r.Use(
		middleware.Recovery(log),
		middleware.Logger(log),
		middleware.CORS(),
	)
	RegisterRoutes(r,
		authHandler,
		jwtServices,
		userRepo,
		blacklist,
		rateLimiter,
		profileHandler,
		addressHandler,
		adminHandler,
		categoryHandler,
		brandHandler,
		productHandler,
		productVariantHandler,
		productImageHandler,
		storeProductHandler,
		cartHandler,
		wishlistHandler,
		notificationHandler,
		couponHandler,
		orderHandler,
		shipmentHandler,
		returnRequestHandler,
	)
	return r
}
