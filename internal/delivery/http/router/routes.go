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
	categoryHandler *handler.CategoryHandler,
	brandHandler *handler.BrandHandler,
) {

	v1 := r.Group("/api/v1")
	v1.GET("/health", handler.Health)
	authMiddleware := middleware.AuthMiddleware(jwtService, userRepo, blacklist)
	auth := v1.Group("/auth")
	protected := auth.Group("")
	protected.Use(authMiddleware)
	{
		//user authentication
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
		//user profile
		protected.GET("/profile", profileHandler.GetProfile)
		protected.PUT("/profile", middleware.RateLimit(rateLimiter, 10, 60, "profile-update"), profileHandler.UpdateProfile)
		protected.PUT("/profile/password", middleware.RateLimit(rateLimiter, 5, 300, "change-password"), profileHandler.ChangePassword)
		protected.POST("/profile/email", middleware.RateLimit(rateLimiter, 3, 300, "email-change-send"), profileHandler.SendEmailChangeOTP)
		protected.POST("/profile/email/verify", middleware.RateLimit(rateLimiter, 5, 300, "email-change-verify"), profileHandler.VerifyEmailChangeOTP)
		protected.POST("/profile/image", middleware.RateLimit(rateLimiter, 5, 300, "profile-image-upload"), profileHandler.UploadProfileImage)
		protected.DELETE("/profile/image", middleware.RateLimit(rateLimiter, 5, 300, "profile-image-delete"), profileHandler.DeleteProfileImage)

		//user addresses
		protected.GET("/addresses", addressHandler.GetAddresses)
		protected.POST("/addresses", middleware.RateLimit(rateLimiter, 20, 60, "address-add"), addressHandler.AddAddress)
		protected.PUT("/addresses/:id", middleware.RateLimit(rateLimiter, 20, 60, "address-update"), addressHandler.UpdateAddress)
		protected.DELETE("/addresses/:id", middleware.RateLimit(rateLimiter, 20, 60, "address-delete"), addressHandler.DeleteAddress)
		protected.PATCH("/addresses/:id/default", middleware.RateLimit(rateLimiter, 20, 60, "address-default"), addressHandler.SetDefaultAddress)
	}
	auth = v1.Group("/admin/auth")
	protected = auth.Group("")
	adminOnly := middleware.AdminOnly()
	protected.Use(authMiddleware, adminOnly)
	{
		//admin authentication
		auth.POST("/login", middleware.RateLimit(rateLimiter, 5, 60, "admin-login"), authHandler.AdminLogin)
		auth.POST("/google", middleware.RateLimit(rateLimiter, 5, 60, "admin-google-login"), authHandler.GoogleLogin)
		auth.POST("/refresh-token", middleware.RateLimit(rateLimiter, 10, 60, "admin-refresh-token"), authHandler.RefreshToken)
		auth.POST("/forgot-password", middleware.RateLimit(rateLimiter, 3, 300, "admin-forgot-password"), authHandler.ForgotPassword)
		auth.POST("/forgot-password/verify-otp", middleware.RateLimit(rateLimiter, 5, 300, "admin-forgot-password-verify"), authHandler.VerifyForgotPassword)
		auth.POST("/forgot-password/resend-otp", middleware.RateLimit(rateLimiter, 3, 300, "admin-forgot-password-resend"), authHandler.ResendForgotPasswordOTP)
		auth.POST("/reset-password", middleware.RateLimit(rateLimiter, 5, 600, "admin-reset-password"), authHandler.ResetPassword)

		protected.POST("/logout", middleware.RateLimit(rateLimiter, 10, 60, "admin-logout"), authHandler.Logout)
	}
	profile = v1.Group("/admin")
	protected = profile.Group("")
	protected.Use(authMiddleware, adminOnly)
	{
		//admin profile
		protected.GET("/profile", profileHandler.GetProfile)
		protected.PUT("/profile", middleware.RateLimit(rateLimiter, 10, 60, "admin-profile-update"), profileHandler.UpdateProfile)
		protected.PUT("/profile/password", middleware.RateLimit(rateLimiter, 5, 300, "admin-change-password"), profileHandler.ChangePassword)
		protected.POST("/profile/email", middleware.RateLimit(rateLimiter, 3, 300, "admin-email-change-send"), profileHandler.SendEmailChangeOTP)
		protected.POST("/profile/email/verify", middleware.RateLimit(rateLimiter, 5, 300, "admin-email-change-verify"), profileHandler.VerifyEmailChangeOTP)
		protected.POST("/profile/image", middleware.RateLimit(rateLimiter, 5, 300, "admin-profile-image-upload"), profileHandler.UploadProfileImage)
		protected.DELETE("/profile/image", middleware.RateLimit(rateLimiter, 5, 300, "admin-profile-image-delete"), profileHandler.DeleteProfileImage)

		//admin address
		protected.GET("/addresses", addressHandler.GetAddresses)
		protected.POST("/addresses", middleware.RateLimit(rateLimiter, 20, 60, "admin-address-add"), addressHandler.AddAddress)
		protected.PUT("/addresses/:id", middleware.RateLimit(rateLimiter, 20, 60, "admin-address-update"), addressHandler.UpdateAddress)
		protected.DELETE("/addresses/:id", middleware.RateLimit(rateLimiter, 20, 60, "admin-address-delete"), addressHandler.DeleteAddress)
		protected.PATCH("/addresses/:id/default", middleware.RateLimit(rateLimiter, 20, 60, "admin-address-default"), addressHandler.SetDefaultAddress)
	}
	admin := v1.Group("/admin")
	admin.Use(authMiddleware, adminOnly)
	{
		//user managment
		admin.GET("/users", adminHandler.ListUsers)
		admin.GET("/users/:id", adminHandler.GetCustomer)
		admin.POST("/users", middleware.RateLimit(rateLimiter, 20, 60, "admin-user-add"), adminHandler.AddCustomer)
		admin.PATCH("/users/:id/block", middleware.RateLimit(rateLimiter, 30, 60, "admin-user-block"), adminHandler.BlockUser)
		admin.PATCH("/users/:id/unblock", middleware.RateLimit(rateLimiter, 30, 60, "admin-user-unblock"), adminHandler.UnblockUser)
		admin.PUT("/users/:id", middleware.RateLimit(rateLimiter, 30, 60, "admin-user-update"), adminHandler.UpdateCustomer)
		admin.DELETE("/users/:id", middleware.RateLimit(rateLimiter, 10, 60, "admin-user-delete"), adminHandler.DeleteCustomer)

		//categories
		admin.POST("/categories", categoryHandler.CreateCategory)
		admin.GET("/categories", categoryHandler.ListCategories)
		admin.GET("/categories/:id", categoryHandler.GetCategory)
		admin.PUT("/categories/:id", categoryHandler.UpdateCategory)
		admin.DELETE("/categories/:id", categoryHandler.DeleteCategory)

		//deleted categories
		admin.GET("/categories/deleted", categoryHandler.ListDeletedCategories)
		admin.PATCH("/categories/:id/restore", categoryHandler.RestoreCategory)
		admin.PATCH("/categories/:id/toggle", categoryHandler.ToggleCategoryActive)

		//brands
		admin.POST("/brands", brandHandler.CreateBrand)
		admin.GET("/brands", brandHandler.ListBrands)
		admin.GET("/brands/:id", brandHandler.GetBrand)
		admin.PUT("/brands/:id", brandHandler.UpdateBrand)
		admin.DELETE("/brands/:id", brandHandler.DeleteBrand)
		
		admin.PUT("/brands/:id/logo", brandHandler.UpdateBrandLogo)
		admin.DELETE("/brands/:id/logo", brandHandler.DeleteBrandLogo)

		//deleted brands
		admin.GET("/brands/deleted", brandHandler.ListDeletedBrands)
		admin.PATCH("/brands/:id/restore", brandHandler.RestoreBrand)
		admin.PATCH("/brands/:id/toggle", brandHandler.ToggleBrandActive)
	}
}
