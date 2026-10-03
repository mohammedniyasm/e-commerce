package router

import (
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/middleware"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine,
	authHandler *handler.AuthHandler,
	jwtService interfaces.JWTService,
	userRepo interfaces.UserRepository,
	blacklist interfaces.AccessTokenBlacklist,
	rateLimiter interfaces.RateLimiter,
	profileHandler *handler.ProfileHandler,
	addressHandler *handler.AddressHandler,
	adminHandler *handler.AdminHandler,
) {

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

		protected.POST("/logout", authHandler.Logout)
		protected.POST("/email/send-verification-otp", middleware.RateLimit(rateLimiter, 3, 300, "email-verification-send"), authHandler.SendVerificationOTP)
		protected.POST("/email/verify-email", middleware.RateLimit(rateLimiter, 5, 300, "email-verification-verify"), authHandler.VerifyEmail)
		protected.POST("/email/resend-otp", middleware.RateLimit(rateLimiter, 3, 300, "email-verification-resend"), authHandler.ResendVerificationOTP)
	}
	profile := v1.Group("/user")
	protected = profile.Group("")
	protected.Use(authMiddleware)
	{
		protected.GET("/profile", profileHandler.GetProfile)
		protected.PUT("/profile", profileHandler.UpdateProfile)
		protected.PUT("/profile/password", profileHandler.ChangePassword)
		protected.POST("/profile/email", profileHandler.SendEmailChangeOTP)
		protected.POST("/profile/email/verify", profileHandler.VerifyEmailChangeOTP)
		protected.POST("/profile/image", profileHandler.UploadProfileImage)
		protected.DELETE("/profile/image", profileHandler.DeleteProfileImage)

		protected.GET("/addresses", addressHandler.GetAddresses)
		protected.POST("/addresses", addressHandler.AddAddress)
		protected.PUT("/addresses/:id", addressHandler.UpdateAddress)
		protected.DELETE("/addresses/:id", addressHandler.DeleteAddress)
		protected.PATCH("/addresses/:id/default", addressHandler.SetDefaultAddress)
	}
	auth = v1.Group("/admin/auth")
	protected = auth.Group("")
	adminOnly := middleware.AdminOnly()
	protected.Use(authMiddleware, adminOnly)
	{
		auth.POST("/login", middleware.RateLimit(rateLimiter, 5, 60, "login"), authHandler.AdminLogin)
		auth.POST("/google", middleware.RateLimit(rateLimiter, 5, 60, "google-login"), authHandler.GoogleLogin)
		auth.POST("/refresh-token", middleware.RateLimit(rateLimiter, 10, 60, "refresh-token"), authHandler.RefreshToken)
		auth.POST("/forgot-password", middleware.RateLimit(rateLimiter, 3, 300, "forgot-password"), authHandler.ForgotPassword)
		auth.POST("/forgot-password/verify-otp", middleware.RateLimit(rateLimiter, 5, 300, "forgot-password-verify"), authHandler.VerifyForgotPassword)
		auth.POST("/forgot-password/resend-otp", middleware.RateLimit(rateLimiter, 3, 300, "forgot-password-resend"), authHandler.ResendForgotPasswordOTP)
		auth.POST("/reset-password", middleware.RateLimit(rateLimiter, 5, 600, "reset-password"), authHandler.ResetPassword)

		protected.POST("/logout", authHandler.Logout)
		// protected.POST("/email/send-verification-otp", middleware.RateLimit(rateLimiter, 3, 300, "email-verification-send"), authHandler.SendVerificationOTP)
		// protected.POST("/email/verify-email", middleware.RateLimit(rateLimiter, 5, 300, "email-verification-verify"), authHandler.VerifyEmail)
		// protected.POST("/email/resend-otp", middleware.RateLimit(rateLimiter, 3, 300, "email-verification-resend"), authHandler.ResendVerificationOTP)
	}
	profile = v1.Group("/admin")
	protected = profile.Group("")
	protected.Use(authMiddleware, adminOnly)
	{
		protected.GET("/profile", profileHandler.GetProfile)
		protected.PUT("/profile", profileHandler.UpdateProfile)
		protected.PUT("/profile/password", profileHandler.ChangePassword)
		protected.POST("/profile/email", profileHandler.SendEmailChangeOTP)
		protected.POST("/profile/email/verify", profileHandler.VerifyEmailChangeOTP)
		protected.POST("/profile/image", profileHandler.UploadProfileImage)
		protected.DELETE("/profile/image", profileHandler.DeleteProfileImage)

		protected.GET("/addresses", addressHandler.GetAddresses)
		protected.POST("/addresses", addressHandler.AddAddress)
		protected.PUT("/addresses/:id", addressHandler.UpdateAddress)
		protected.DELETE("/addresses/:id", addressHandler.DeleteAddress)
		protected.PATCH("/addresses/:id/default", addressHandler.SetDefaultAddress)
	}
	admin := v1.Group("/admin")
	admin.Use(authMiddleware, adminOnly)
	{
		admin.GET("/users", adminHandler.ListUsers)
		admin.GET("/users/:id", adminHandler.GetCustomer)
		admin.POST("/users", adminHandler.AddCustomer)
		admin.PATCH("/users/:id/block", adminHandler.BlockUser)
		admin.PATCH("/users/:id/unblock", adminHandler.UnblockUser)
		admin.PUT("/users/:id", adminHandler.UpdateCustomer)
		admin.DELETE("/users/:id", adminHandler.DeleteCustomer)
	}
}
