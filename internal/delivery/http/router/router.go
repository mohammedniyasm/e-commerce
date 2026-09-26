package router

import (
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/middleware"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func SetupRouter(log *slog.Logger, authHandler *handler.AuthHandler) *gin.Engine {
	r := gin.Default()
	r.Use(
		middleware.Recovery(log),
		middleware.Logger(log),
		middleware.CORS(),
	)
	RegisterRoutes(r, authHandler)
	return r
}
