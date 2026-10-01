package router

import (
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/middleware"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler, jwtService interfaces.JWTService, userRepo interfaces.UserRepository, blacklist interfaces.AccessTokenBlacklist) {
	v1 := r.Group("/api/v1")
	v1.GET("/health", handler.Health)
	authMiddleware := middleware.AuthMiddleware(jwtService, userRepo, blacklist)
	auth := v1.Group("/auth")
	protected := auth.Group("")
	protected.Use(authMiddleware)
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh-token", authHandler.RefreshToken)
		// auth.POST("/logout", authHandler.Logout)
		protected.POST("/logout", authHandler.Logout)
		protected.POST("/email/send-verification-otp", authHandler.SendVerificationOTP)
		protected.POST("/email/verify-email", authHandler.VerifyEmail)
		protected.POST("/email/resend-otp", authHandler.ResendVerificationOTP)
	}
}
