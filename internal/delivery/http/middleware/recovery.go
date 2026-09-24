package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func Recovery(log *slog.Logger)gin.HandlerFunc{
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("panic recovered",
			"error",recovered,
			"path",c.Request.URL.Path,
			"method",c.Request.Method,
	)
	c.AbortWithStatusJSON(500,gin.H{
		"error":"internal server error",
	})
	})
}