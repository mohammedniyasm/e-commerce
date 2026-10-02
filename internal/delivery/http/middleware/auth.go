package middleware

import (
	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/usecase/interfaces"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService interfaces.JWTService, userRepo interfaces.UserRepository,blacklist interfaces.AccessTokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, response.APIResponse{
				Success: false,
				Message: "authorization header required",
			})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, response.APIResponse{
				Success: false,
				Message: "invalid authorization header",
			})
			return
		}
		claims, err := jwtService.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(401, response.APIResponse{
				Success: false,
				Message: "invalid or expired access token",
			})
			return
		}
		blacklisted,err:=blacklist.IsBlacklisted(c.Request.Context(),claims.JTI)
		if err != nil{
			c.AbortWithStatusJSON(401,response.APIResponse{
				Success: false,
				Message: "failed to verify access token",
			})
			return 
		}
		if blacklisted{
			c.AbortWithStatusJSON(401,response.APIResponse{
				Success: false,
				Message: "access token has been revoked",
			})
			return 
		}
		userID, err := strconv.ParseUint(claims.UserID, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(401, response.APIResponse{
				Success: false,
				Message: "invalid user identity",
			})
			return
		}
		blocked, err := userRepo.IsBlocked(c.Request.Context(), uint(userID))
		if err != nil {
			c.AbortWithStatusJSON(500, response.APIResponse{
				Success: false,
				Message: "failed to verify account status",
			})
			return
		}
		if blocked {
			c.AbortWithStatusJSON(403, response.APIResponse{
				Success: false,
				Message: "user account is blocked",
			})
			return
		}
		c.Set(UserIdKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Set(AccessClaimsKey,claims)
		c.Next()
	}
}
