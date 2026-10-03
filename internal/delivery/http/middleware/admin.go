package middleware

import (
	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/domain/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get(UserRoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, response.APIResponse{
				Success: false,
				Message: "user role not found",
			})
			return
		}
		role, ok := roleValue.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, response.APIResponse{
				Success: false,
				Message: "invalid user role",
			})
			return
		}
		if role != string(models.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, response.APIResponse{
				Success: false,
				Message: "admin access required",
			})
			return
		}
		c.Next()
	}
}
