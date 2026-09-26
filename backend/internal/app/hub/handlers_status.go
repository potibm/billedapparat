package hub

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

const (
	// readinessCheckTimeout bounds the probe so a stalled pool cannot hang it.
	readinessCheckTimeout = 2 * time.Second

	// readinessReportInterval throttles Sentry events; every failure is still logged.
	readinessReportInterval = time.Minute
)

// handleGetHealth is the liveness probe; it performs no dependency checks.
func (s *Server) handleGetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleGetReady is the readiness probe, see ADR 007.
func (s *Server) handleGetReady(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), readinessCheckTimeout)
	defer cancel()

	if err := s.pinger.Ping(ctx); err != nil {
		s.reportReadinessFailure(c, err)

		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "database down"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// reportReadinessFailure logs every failed probe and throttles Sentry reports.
func (s *Server) reportReadinessFailure(c *gin.Context, pingErr error) {
	s.logger.WarnContext(c.Request.Context(), "Readiness check failed", "error", pingErr)

	now := time.Now().UnixNano()
	last := s.lastReadinessSent.Load()

	if last != 0 && now-last < int64(readinessReportInterval) {
		return
	}

	if !s.lastReadinessSent.CompareAndSwap(last, now) {
		return
	}

	sentryHub := sentry.GetHubFromContext(c.Request.Context())
	if sentryHub == nil {
		// Fall back like the sentrygin middleware: WithScope dereferences its receiver.
		sentryHub = sentry.CurrentHub()
	}

	sentryHub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelWarning)
		scope.SetTag("probe", "readiness")
		sentryHub.CaptureException(pingErr)
	})
}
