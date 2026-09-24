package router

import (
	"ecommerce/internal/delivery/http/middleware"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func SetupRouter(log *slog.Logger) *gin.Engine {
	r:=gin.New()
	r.Use(
		middleware.Recovery(log),
		middleware.Logger(log),
		middleware.CORS(),
	)
	RegisterRoutes(r)
	return r
}