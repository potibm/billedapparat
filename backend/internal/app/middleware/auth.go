package middleware

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/potibm/billedapparat/internal/app/domain"
	sloggin "github.com/samber/slog-gin"
)

const (
	defaultTimeout              = 10 * time.Second
	defaultJWKSRefreshInterval  = 1 * time.Hour
	defaultJWKSRefreshRateLimit = 5 * time.Minute
)

// OIDCDiscovery represents the necessary fields from the /.well-known/openid-configuration.
type OIDCDiscovery struct {
	JwksURI string `json:"jwks_uri"`
	Issuer  string `json:"issuer"`
}

// AuthMiddleware creates a gin handler to verify JWT access tokens.
// clientID is kept in the signature for backwards compatibility, though not strictly checked for access tokens.
func AuthMiddleware(ctx context.Context, issuerURL, clientID string, skipTLSVerify bool) (gin.HandlerFunc, error) {
	client, err := buildHTTPClient(skipTLSVerify)
	if err != nil {
		return nil, err
	}

	jwks, expectedIssuer, err := initJWKS(ctx, client, issuerURL)
	if err != nil {
		return nil, err
	}

	return func(c *gin.Context) {
		reqLogger := slog.With(
			"ip", c.ClientIP(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		tokenString, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			reqLogger.Warn("Invalid authorization header", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})

			return
		}

		userID, err := validateTokenAndGetUserID(tokenString, jwks, expectedIssuer)
		if err != nil {
			reqLogger.Warn("Token verification failed", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})

			return
		}

		// Add user ID to context and logging
		c.Set("userID", userID)
		sloggin.AddCustomAttributes(c, slog.String("user_id", userID))

		ctxWithUser := context.WithValue(c.Request.Context(), domain.UserIDKey, userID)
		c.Request = c.Request.WithContext(ctxWithUser)

		c.Next()
	}, nil
}

// buildHTTPClient configures an HTTP client with an optional TLS verification skip.
func buildHTTPClient(skipTLSVerify bool) (*http.Client, error) {
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok || baseTransport == nil {
		return nil, errors.New("default HTTP transport is not *http.Transport")
	}

	transport := baseTransport.Clone()
	if skipTLSVerify {
		// #nosec G402 -- for local dev environments
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // NOSONAR
	}

	return &http.Client{
		Transport: transport,
		Timeout:   defaultTimeout,
	}, nil
}

// initJWKS fetches the OIDC discovery document and initializes the JWKS keyfunc.
func initJWKS(ctx context.Context, client *http.Client, issuerURL string) (*keyfunc.JWKS, string, error) {
	discoveryURL := strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, http.NoBody)
	if err != nil {
		return nil, "", fmt.Errorf("error creating discovery request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching OIDC discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("OIDC discovery returned status: %s", resp.Status)
	}

	var discovery OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, "", fmt.Errorf("error decoding OIDC discovery document: %w", err)
	}

	// Initialize keyfunc which will automatically fetch and cache the JWKS
	options := keyfunc.Options{
		Client: client,
		Ctx:    ctx,
		RefreshErrorHandler: func(err error) {
			slog.Error("Error refreshing JWKS", "error", err)
		},
		RefreshInterval:   defaultJWKSRefreshInterval,
		RefreshUnknownKID: true,
		RefreshRateLimit:  defaultJWKSRefreshRateLimit,
	}

	jwks, err := keyfunc.Get(discovery.JwksURI, options)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get JWKS: %w", err)
	}

	return jwks, discovery.Issuer, nil
}

// extractBearerToken parses the Authorization header and returns the token string.
func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", errors.New("invalid token format")
	}

	return parts[1], nil
}

// validateTokenAndGetUserID validates the JWT signature, checks claims, and extracts the subject.
func validateTokenAndGetUserID(tokenString string, jwks *keyfunc.JWKS, expectedIssuer string) (string, error) {
	token, err := jwt.Parse(tokenString, jwks.Keyfunc)
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims structure")
	}

	iss, ok := claims["iss"].(string)
	if !ok || iss != expectedIssuer {
		return "", errors.New("invalid issuer")
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return "", errors.New("missing subject in token")
	}

	return userID, nil
}
