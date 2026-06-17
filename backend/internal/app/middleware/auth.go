package middleware

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/domain"
	sloggin "github.com/samber/slog-gin"
)

func AuthMiddleware(ctx context.Context, issuerURL, clientID string, skipTLSVerify bool) (gin.HandlerFunc, error) {
	// 1. HTTP client with optional TLS verification
	const oidcHTTPTimeout = 10 * time.Second

	baseTransport, _ := http.DefaultTransport.(*http.Transport)
	transport := baseTransport.Clone()
	client := &http.Client{
		Transport: transport,
		Timeout:   oidcHTTPTimeout,
	}

	if skipTLSVerify {
		// #nosec G402 -- for local dev environments
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // NOSONAR
	}

	// 2. Add the custom HTTP client to the OIDC context
	setupCtx, cancel := context.WithTimeout(ctx, oidcHTTPTimeout)
	defer cancel()

	oidcCtx := oidc.ClientContext(setupCtx, client)

	// 3. Initialize the OIDC Provider
	provider, err := oidc.NewProvider(oidcCtx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("error initializing the OIDC Provider: %w", err)
	}

	// 4. Configure the verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	return func(c *gin.Context) {
		reqLogger := slog.With(
			"ip", c.ClientIP(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			reqLogger.Warn("Missing Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})

			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			reqLogger.Warn("Invalid token format", "header_length", len(parts))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})

			return
		}

		tokenString := parts[1]

		reqCtx := oidc.ClientContext(c.Request.Context(), client)

		idToken, err := verifier.Verify(reqCtx, tokenString)
		if err != nil {
			reqLogger.Warn("Token verification failed", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})

			return
		}

		userID := idToken.Subject

		c.Set("userID", userID)
		sloggin.AddCustomAttributes(c, slog.String("user_id", userID))

		ctxWithUser := context.WithValue(c.Request.Context(), domain.UserIDKey, userID)
		c.Request = c.Request.WithContext(ctxWithUser)

		c.Next()
	}, nil
}
