package middleware

import (
	"fmt"
	"strconv"

	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/usecase/interfaces"

	"github.com/gin-gonic/gin"
)

func RateLimit(limiter interfaces.RateLimiter, limit int, windowSeconds int, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := c.ClientIP()
		if userID, exists := c.Get(UserIdKey); exists {
			identity = fmt.Sprintf("user:%v", userID)
		} else {
			identity = "ip:" + identity
		}
		key := fmt.Sprintf("%s:%s", prefix, identity)
		allowed, err := limiter.Allow(c.Request.Context(), key, limit, windowSeconds)
		if err != nil {
			c.AbortWithStatusJSON(503, response.APIResponse{
				Success: false,
				Message: "rate limit service unavailable",
			},
			)
			return
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(windowSeconds))
			c.AbortWithStatusJSON(429, response.APIResponse{
				Success: false,
				Message: "too many requests, please try again later",
			},
			)
			return
		}
		c.Next()
	}
}
