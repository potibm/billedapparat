package hub

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPinger implements repository.Pinger for testing.
type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

func TestHandleGetHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := &Server{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", http.NoBody)

	s.handleGetHealth(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
}

func TestHandleGetReady(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("database reachable", func(t *testing.T) {
		s := &Server{
			pinger: &mockPinger{},
			logger: newTestLoggerForHandler(),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/ready", http.NoBody)

		s.handleGetReady(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var body map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "ready", body["status"])
		assert.NotContains(t, body, "error")
	})

	t.Run("database unreachable", func(t *testing.T) {
		s := &Server{
			pinger: &mockPinger{err: errors.New("sql: database is closed")},
			logger: newTestLoggerForHandler(),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/ready", http.NoBody)

		s.handleGetReady(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var body map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "unavailable", body["status"])
		assert.Equal(t, "database down", body["error"])
	})
}

func TestHandleGetReadyCancelledRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := &Server{
		pinger: &mockPinger{err: context.Canceled},
		logger: newTestLoggerForHandler(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", http.NoBody).WithContext(ctx)

	s.handleGetReady(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestReportReadinessFailureThrottling also guards the nil-hub path.
func TestReportReadinessFailureThrottling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	failure := errors.New("sql: database is closed")

	s := &Server{logger: newTestLoggerForHandler()}

	newContext := func() *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/ready", http.NoBody)

		return c
	}

	// First failure is reported and stamps the window.
	s.reportReadinessFailure(newContext(), failure)
	assert.NotZero(t, s.lastReadinessSent.Load())

	stamped := s.lastReadinessSent.Load()

	// Further failures inside the window are suppressed.
	for range 5 {
		s.reportReadinessFailure(newContext(), failure)
	}

	assert.Equal(t, stamped, s.lastReadinessSent.Load())

	// Once the window has passed a later failure reports again.
	s.lastReadinessSent.Store(time.Now().Add(-2 * readinessReportInterval).UnixNano())
	s.reportReadinessFailure(newContext(), failure)
	assert.Greater(t, s.lastReadinessSent.Load(), stamped)
}
