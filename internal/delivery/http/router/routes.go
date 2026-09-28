package router

import (
	"ecommerce/internal/delivery/http/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler) {
	v1 := r.Group("/api/v1")
	v1.GET("/health", handler.Health)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login",authHandler.Login)
		auth.POST("/refresh-token",authHandler.RefreshToken)
	}
}
