package bootstrap

import (
	"ecommerce/config"
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/router"
	authinfra "ecommerce/internal/infrastructure/auth"
	"ecommerce/internal/infrastructure/database"
	"ecommerce/internal/infrastructure/email"
	"ecommerce/internal/infrastructure/invoice"
	"ecommerce/internal/infrastructure/redis"
	"ecommerce/internal/infrastructure/storage"
	"ecommerce/internal/repository/postgres"
	"ecommerce/internal/usecase/address"
	"ecommerce/internal/usecase/admin"
	"ecommerce/internal/usecase/auth"
	brand "ecommerce/internal/usecase/brands"
	"ecommerce/internal/usecase/cart"
	"ecommerce/internal/usecase/category"
	couponuseCase "ecommerce/internal/usecase/coupons"
	"ecommerce/internal/usecase/notification"
	"ecommerce/internal/usecase/order"
	"ecommerce/internal/usecase/productimage"
	"ecommerce/internal/usecase/products"
	"ecommerce/internal/usecase/productvariant"
	"ecommerce/internal/usecase/profile"
	return_request "ecommerce/internal/usecase/returnrequest"
	"ecommerce/internal/usecase/shipment"
	"ecommerce/internal/usecase/wishlist"
	"ecommerce/internal/worker"
	"ecommerce/internal/worker/coupon"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Application struct {
	Router        *gin.Engine
	WorkerManager *worker.Manager
}

func NewApplication(db *gorm.DB, log *slog.Logger, cfg config.Config) (*Application, error) {
	redisClient := redis.NewRedisClient(cfg.Redis, log)
	rateLimiter := redis.NewRedisRateLimiter(redisClient)

	objectStorage, err := storage.NewMinIOStorage(storage.MinIOConfig(cfg.MinIO))
	if err != nil {
		return nil, fmt.Errorf("initialize MinIO storage: %w", err)
	}

	accessTokenBlacklistStore := redis.NewAccessTokenBlacklistStore(redisClient)
	refreshSessionStore := redis.NewRefreshSessionStore(redisClient, cfg.JWT.RefreshExpiry)

	otpStore := redis.NewOTPStore(redisClient)
	jwtService := authinfra.NewJWTService(cfg.JWT)

	emailSender := email.NewSMTPEmailSender(cfg.Email)
	googleTokenVerifier := authinfra.NewGoogleTokenVerifier(cfg.Google.ClientID)

	userRepository := postgres.NewUserRepository(db)
	authUseCase := auth.NewAuthUseCase(userRepository, *otpStore, jwtService, refreshSessionStore, accessTokenBlacklistStore, emailSender, googleTokenVerifier, log)
	authHandler := handler.NewAuthHandler(authUseCase, cfg.Cookie)

	profileUsecase := profile.NewProfileUseCase(userRepository, otpStore, emailSender, objectStorage, log)
	profileHandler := handler.NewProfileHandler(profileUsecase)

	addressRepository := postgres.NewAddressRepository(db)
	addressUseCase := address.NewAddressUseCase(addressRepository, log)
	addressHandler := handler.NewAddressHandler(addressUseCase)

	adminUseCase := admin.NewAdminUseCase(userRepository, log)
	adminHandler := handler.NewAdminHandler(adminUseCase)

	productRepository := postgres.NewProductRepository(db)

	categoryRepository := postgres.NewCategoryRepository(db)
	categoryUseCase := category.NewCategoryUseCase(categoryRepository, productRepository, log)
	categoryHandler := handler.NewCategoryHandler(categoryUseCase)

	brandRepository := postgres.NewBrandRepository(db)
	brandUseCase := brand.NewBrandUseCase(brandRepository, productRepository, objectStorage, log)
	brandHandler := handler.NewBrandHandler(brandUseCase)

	productUseCase := products.NewProductUseCase(productRepository, categoryRepository, brandRepository, log)
	productHandler := handler.NewProductHandler(productUseCase)

	productVariantRepository := postgres.NewProductVariantRepository(db)
	productVariantUseCase := productvariant.NewProductVariantUseCase(productVariantRepository, productRepository, rateLimiter, log)
	productVariantHandler := handler.NewProductVariantHandler(productVariantUseCase)

	productImageRepository := postgres.NewProductImageRepository(db)
	productImageUseCase := productimage.NewProductImageUseCase(productImageRepository, productRepository, productVariantRepository, objectStorage, log)
	productImageHandler := handler.NewProductImageHandler(productImageUseCase)

	cartRepository := postgres.NewCartRepository(db)
	cartUseCase := cart.NewCartUseCase(cartRepository, productVariantRepository, log)
	cartHandler := handler.NewCartHandler(cartUseCase)

	wishlistRepository := postgres.NewWishlistRepository(db)
	wishlistUseCase := wishlist.NewWishlistUseCase(wishlistRepository, log)
	wishlistHandler := handler.NewWishlistHandler(wishlistUseCase)

	pdfGen := invoice.NewPDFGenerator("Drape")
	orderItemRepository := postgres.NewOrderItemRepository(db)
	orderRepository := postgres.NewOrderRepository(db)
	unitOfWork := database.NewGormUnitOfWork(db)
	orderUsecase := order.NewOrderUseCase(orderRepository, unitOfWork, pdfGen, log)
	orderHandler := handler.NewOrderHandler(orderUsecase, log)

	notificationRepo := postgres.NewNotificationRepository(db)
	notificationUseCase := notification.NewNotificationUseCase(notificationRepo, unitOfWork, log)
	notificationHandler := handler.NewNotificationHandler(notificationUseCase, log)

	workerManager := worker.NewManager(log)

	couponRepository := postgres.NewCouponRepository(db)
	couponUsageRepository := postgres.NewCouponUsageRepository(db)
	couponUsage := couponuseCase.NewCouponUseCase(couponRepository, orderRepository, couponUsageRepository, log)
	couponHandler := handler.NewCouponHandler(couponUsage, log)

	couponWorker := coupon.NewWorker(couponRepository, notificationRepo, couponUsageRepository, notificationUseCase, time.Minute, log)

	shipmentRepository := postgres.NewShipmentRepository(db)
	shipmentUseCase := shipment.NewShipmentUseCase(shipmentRepository, orderRepository, unitOfWork, log)
	shipmentHandler := handler.NewShipmentHandler(shipmentUseCase, log)

	returnRequestRepository := postgres.NewReturnRequestRepository(db)
	returnRequestUseCase := return_request.NewReturnRequestUseCase(returnRequestRepository, orderRepository, orderItemRepository, productVariantRepository, unitOfWork, log)
	returnRequestHandler := handler.NewReturnRequestHandler(returnRequestUseCase)

	workerManager.Register(couponWorker)

	storeProductHandler := handler.NewStoreProductHandler(productUseCase)
	r := router.SetupRouter(log,
		authHandler,
		jwtService,
		userRepository,
		accessTokenBlacklistStore,
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

	return &Application{
		Router:        r,
		WorkerManager: workerManager,
	}, nil
}
