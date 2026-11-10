package bootstrap

import (
	"ecommerce/config"
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/router"
	authinfra "ecommerce/internal/infrastructure/auth"
	"ecommerce/internal/infrastructure/database"
	"ecommerce/internal/infrastructure/email"
	"ecommerce/internal/infrastructure/invoice"
	"ecommerce/internal/infrastructure/payment/razorpay"
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
	"ecommerce/internal/usecase/dashboard"
	"ecommerce/internal/usecase/notification"
	"ecommerce/internal/usecase/order"
	"ecommerce/internal/usecase/payment"
	"ecommerce/internal/usecase/productimage"
	"ecommerce/internal/usecase/products"
	"ecommerce/internal/usecase/productvariant"
	"ecommerce/internal/usecase/profile"
	"ecommerce/internal/usecase/referral"
	"ecommerce/internal/usecase/report"
	return_request "ecommerce/internal/usecase/returnrequest"
	"ecommerce/internal/usecase/review"
	"ecommerce/internal/usecase/shipment"
	"ecommerce/internal/usecase/wallet"
	"ecommerce/internal/usecase/wishlist"
	"ecommerce/internal/worker"
	"ecommerce/internal/worker/coupon"
	orderworker "ecommerce/internal/worker/order"
	referralworker "ecommerce/internal/worker/referral"
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
	unitOfWork := database.NewGormUnitOfWork(db)
	accessTokenBlacklistStore := redis.NewAccessTokenBlacklistStore(redisClient)
	refreshSessionStore := redis.NewRefreshSessionStore(redisClient, cfg.JWT.RefreshExpiry)

	otpStore := redis.NewOTPStore(redisClient)
	jwtService := authinfra.NewJWTService(cfg.JWT)

	emailSender := email.NewSMTPEmailSender(cfg.Email)
	googleTokenVerifier := authinfra.NewGoogleTokenVerifier(cfg.Google.ClientID)

	userRepository := postgres.NewUserRepository(db)
	authUseCase := auth.NewAuthUseCase(userRepository, otpStore, jwtService, refreshSessionStore, accessTokenBlacklistStore, emailSender, googleTokenVerifier, unitOfWork, log)
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

	walletTopUpRepository := postgres.NewWalletTopUpRepository(db)
	walletRepository := postgres.NewWalletRepository(db)

	shipmentRepository := postgres.NewShipmentRepository(db)
	shipmentUseCase := shipment.NewShipmentUseCase(shipmentRepository, orderRepository, unitOfWork, log)
	shipmentHandler := handler.NewShipmentHandler(shipmentUseCase, log)

	returnRequestRepository := postgres.NewReturnRequestRepository(db)
	returnRequestUseCase := return_request.NewReturnRequestUseCase(returnRequestRepository, orderRepository, orderItemRepository, productVariantRepository, shipmentRepository, unitOfWork, log)
	returnRequestHandler := handler.NewReturnRequestHandler(returnRequestUseCase)

	refferalRepository := postgres.NewReferralRepository(db)
	refferalSettingsRepository := postgres.NewReferralSettingsRepository(db)
	refferalUseCase := referral.NewReferralUseCase(userRepository, refferalRepository, refferalSettingsRepository, log)
	refferalHandler := handler.NewReferralHandler(refferalUseCase, log)

	razorpayClient, err := razorpay.NewClient(cfg.Razorpay.KeyID, cfg.Razorpay.KeySecret)
	if err != nil {
		return nil, err
	}
	paymentRepository := postgres.NewPaymentRepository(db)
	paymentUseCase := payment.NewPaymentUseCase(orderRepository, paymentRepository, razorpayClient, unitOfWork, cfg.Razorpay.KeyID, cfg.Razorpay.WebhookSecret)
	paymentHandler := handler.NewPaymentHandler(paymentUseCase)
	storeProductHandler := handler.NewStoreProductHandler(productUseCase)

	walletUseCase := wallet.NewWalletUseCase(walletRepository, walletTopUpRepository, razorpayClient, unitOfWork, cfg.Razorpay.KeyID, log)
	walletHandler := handler.NewWalletHandler(walletUseCase, cfg.Razorpay.KeyID, cfg.Razorpay.WebhookSecret)

	reviewRepository := postgres.NewReviewRepository(db)
	reviewUseCase := review.NewReviewUseCase(reviewRepository, orderItemRepository, orderRepository, shipmentRepository, productVariantRepository, log)
	reviewHandler := handler.NewReviewHandler(reviewUseCase)

	dashboardRepository := postgres.NewDashboardRepository(db)
	dashboardUseCase := dashboard.NewDashboardUseCase(dashboardRepository, log)
	dashboardHandler := handler.NewDashboardHandler(dashboardUseCase)

	reportRepo := postgres.NewReportRepository(db)
	reportUseCase := report.NewReportUseCase(reportRepo, log)
	reportHandler := handler.NewReportHandler(reportUseCase)

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
		walletHandler,
		refferalHandler,
		paymentHandler,
		reviewHandler,
		dashboardHandler,
		reportHandler,
	)

	couponWorker := coupon.NewWorker(couponRepository, notificationRepo, couponUsageRepository, notificationUseCase, 4*time.Hour, log)
	refferalWorker := referralworker.NewWorker(refferalRepository, unitOfWork, 4*time.Hour, log)
	orderExpiryWorker := orderworker.NewWorker(orderRepository, orderUsecase, 30*time.Minute, log)

	workerManager.Register(couponWorker)
	workerManager.Register(refferalWorker)
	workerManager.Register(orderExpiryWorker)

	return &Application{
		Router:        r,
		WorkerManager: workerManager,
	}, nil
}
