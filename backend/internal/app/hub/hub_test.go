package hub

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleSPAFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	srv := &Server{
		staticFiles: fstest.MapFS{
			"assets/index.html": &fstest.MapFile{Data: []byte("<!DOCTYPE html><html></html>")},
		},
		logger: slog.Default(),
	}

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "API route should be ignored (404 handled by gin)",
			url:            "/api/unknown",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Missing static file should be ignored",
			url:            "/main.js",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "OAuth callback with dots in query should return index.html",
			url:            "/auth-callback?iss=https://id.evoke.eu/auth/v1/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Normal SPA route should return index.html",
			url:            "/admin/dashboard",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, engine := gin.CreateTestContext(w)

			engine.NoRoute(srv.handleSPAFallback)

			req, _ := http.NewRequest(http.MethodGet, tt.url, http.NoBody)
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
