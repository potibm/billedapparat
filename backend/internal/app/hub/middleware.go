package hub

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/config"
)

const (
	collectorSourceKey = "collector_source"
	collectorTypeKey   = "collector_type"
)

func APIKeyAuthMiddleware(validAdminKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)

		if token == "" || token != validAdminKey {
			slog.Warn("Unauthorized admin access attempt", "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})

			return
		}

		c.Next()
	}
}

func CollectorAuthMiddleware(collectors map[string]config.CollectorConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)

		if token == "" {
			slog.Warn("Unauthorized collector access attempt (missing token)", "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})

			return
		}

		var authenticatedSource string

		isAuthenticated := false

		// Check against all configured collectors
		for sourceName, cfg := range collectors {
			if cfg.Enabled && cfg.APIKey == token {
				isAuthenticated = true
				authenticatedSource = sourceName

				break
			}
		}

		if !isAuthenticated {
			slog.Warn("Unauthorized collector access attempt (invalid token)", "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})

			return
		}

		// Store collector name in context for later use in handlers
		c.Set(collectorSourceKey, authenticatedSource)
		c.Set(collectorTypeKey, string(collectors[authenticatedSource].Type))

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	return strings.TrimSpace(token)
}
