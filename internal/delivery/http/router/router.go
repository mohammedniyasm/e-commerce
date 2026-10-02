package router

import (
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/middleware"
	"ecommerce/internal/usecase/interfaces"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func SetupRouter(log *slog.Logger, authHandler *handler.AuthHandler, jwtServices interfaces.JWTService, userRepo interfaces.UserRepository,blacklist interfaces.AccessTokenBlacklist,rateLimiter interfaces.RateLimiter) *gin.Engine {
	r := gin.Default()
	r.Use(
		middleware.Recovery(log),
		middleware.Logger(log),
		middleware.CORS(),
	)
	RegisterRoutes(r, authHandler, jwtServices, userRepo,blacklist,rateLimiter)
	return r
}
