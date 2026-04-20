package hub

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")

		if key == "" || key != validKey {
			slog.Warn("Unauthorized API access attempt", "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})

			return
		}

		c.Next()
	}
}
