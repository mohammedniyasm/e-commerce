package bootstrap

import (
	"ecommerce/config"
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/router"
	authinfra "ecommerce/internal/infrastructure/auth"
	"ecommerce/internal/infrastructure/email"
	"ecommerce/internal/infrastructure/redis"
	"ecommerce/internal/repository/postgres"
	"ecommerce/internal/usecase/auth"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Application struct {
	Router *gin.Engine
}

func NewApplication(db *gorm.DB, log *slog.Logger, cfg config.Config) *Application {
	redisClient := redis.NewRedisClient(cfg.Redis, log)
	accessTokenBlacklistStore:=redis.NewAccessTokenBlacklistStore(redisClient)
	refreshSessionStore := redis.NewRefreshSessionStore(redisClient, cfg.JWT.RefreshExpiry)
	otpStore := redis.NewOTPStore(redisClient)
	jwtService := authinfra.NewJWTService(cfg.JWT)
	emailSender:=email.NewSMTPEmailSender(cfg.Email)
	userRepository := postgres.NewUserRepository(db)
	authUseCase := auth.NewAuthUseCase(userRepository, *otpStore, jwtService, refreshSessionStore,accessTokenBlacklistStore,emailSender, log)
	authHandler := handler.NewAuthHandler(authUseCase,cfg.Cookie)
	
	r := router.SetupRouter(log, authHandler, jwtService, userRepository,accessTokenBlacklistStore)
	return &Application{
		Router: r,
	}
}
