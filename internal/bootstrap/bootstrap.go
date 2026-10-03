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

	profileStorage, err := storage.NewMinIOStorage(storage.MinIOConfig(cfg.MinIO))
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

	profileUsecase := profile.NewProfileUseCase(userRepository, otpStore, emailSender, profileStorage, log)
	profileHandler := handler.NewProfileHandler(profileUsecase)

	addressRepository := postgres.NewAddressRepository(db)
	addressUseCase := address.NewAddressUseCase(addressRepository, log)
	addressHandler := handler.NewAddressHandler(addressUseCase)

	adminUseCase := admin.NewAdminUseCase(userRepository, log)
	adminHandler := handler.NewAdminHandler(adminUseCase)
	r := router.SetupRouter(log,
		authHandler,
		jwtService,
		userRepository,
		accessTokenBlacklistStore,
		rateLimiter,
		profileHandler,
		addressHandler,
		adminHandler,
	)
	return &Application{
		Router: r,
	}, nil
}
