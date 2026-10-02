package router

import (
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/middleware"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler, jwtService interfaces.JWTService, userRepo interfaces.UserRepository, blacklist interfaces.AccessTokenBlacklist, rateLimiter interfaces.RateLimiter) {
	v1 := r.Group("/api/v1")
	v1.GET("/health", handler.Health)
	authMiddleware := middleware.AuthMiddleware(jwtService, userRepo, blacklist)
	auth := v1.Group("/auth")
	protected := auth.Group("")
	protected.Use(authMiddleware)
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", middleware.RateLimit(rateLimiter, 5, 60, "login"), authHandler.Login)
		auth.POST("/google", middleware.RateLimit(rateLimiter, 5, 60, "google-login"), authHandler.GoogleLogin)
		auth.POST("/refresh-token", middleware.RateLimit(rateLimiter, 10, 60, "refresh-token"), authHandler.RefreshToken)
		auth.POST("/forgot-password", middleware.RateLimit(rateLimiter, 3, 300, "forgot-password"), authHandler.ForgotPassword)
		auth.POST("/forgot-password/verify-otp", middleware.RateLimit(rateLimiter, 5, 300, "forgot-password-verify"), authHandler.VerifyForgotPassword)
		auth.POST("/forgot-password/resend-otp", middleware.RateLimit(rateLimiter, 3, 300, "forgot-password-resend"), authHandler.ResendForgotPasswordOTP)
		auth.POST("/reset-password", middleware.RateLimit(rateLimiter, 5, 600, "reset-password"), authHandler.ResetPassword)
		// auth.POST("/logout", authHandler.Logout)
		protected.POST("/logout", authHandler.Logout)
		protected.POST("/email/send-verification-otp", middleware.RateLimit(rateLimiter, 3, 300, "email-verification-send"), authHandler.SendVerificationOTP)
		protected.POST("/email/verify-email", middleware.RateLimit(rateLimiter, 5, 300, "email-verification-verify"), authHandler.VerifyEmail)
		protected.POST("/email/resend-otp", middleware.RateLimit(rateLimiter, 3, 300, "email-verification-resend"), authHandler.ResendVerificationOTP)
	}
}
