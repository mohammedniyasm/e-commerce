package bootstrap

import (
	"ecommerce/config"
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/router"
	authinfra "ecommerce/internal/infrastructure/auth"
	"ecommerce/internal/infrastructure/email"
	"ecommerce/internal/infrastructure/redis"
	"ecommerce/internal/infrastructure/storage"
	"ecommerce/internal/repository/postgres"
	"ecommerce/internal/usecase/address"
	"ecommerce/internal/usecase/admin"
	"ecommerce/internal/usecase/auth"
	brand "ecommerce/internal/usecase/brands"
	"ecommerce/internal/usecase/category"
	"ecommerce/internal/usecase/productimage"
	"ecommerce/internal/usecase/products"
	"ecommerce/internal/usecase/productvariant"
	"ecommerce/internal/usecase/profile"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Application struct {
	Router *gin.Engine
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
	productVariantUseCase := productvariant.NewProductVariantUseCase(productVariantRepository, productRepository, log)
	productVariantHandler := handler.NewProductVariantHandler(productVariantUseCase)

	productImageRepository := postgres.NewProductImageRepository(db)
	productImageUseCase := productimage.NewProductImageUseCase(productImageRepository, productRepository, productVariantRepository, objectStorage, log)
	productImageHandler := handler.NewProductImageHandler(productImageUseCase)

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
	)

	return &Application{
		Router: r,
	}, nil
}
