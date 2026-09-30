package bootstrap

import (
	"ecommerce/config"
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/router"
	authinfra "ecommerce/internal/infrastructure/auth"
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
	refreshSessionStore := redis.NewRefreshSessionStore(redisClient, cfg.JWT.RefreshExpiry)
	otpStore := redis.NewOTPStore(redisClient)
	jwtService := authinfra.NewJWTService(cfg.JWT)
	userRepository := postgres.NewUserRepository(db)
	authUseCase := auth.NewAuthUseCase(userRepository, *otpStore, jwtService, refreshSessionStore, log)
	authHandler := handler.NewAuthHandler(authUseCase)

	r := router.SetupRouter(log, authHandler, jwtService, userRepository)
	return &Application{
		Router: r,
	}
}
