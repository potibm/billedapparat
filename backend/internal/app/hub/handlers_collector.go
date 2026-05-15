package hub

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/config"
)

func (s *Server) validateCollectorAccess(
	ctx *gin.Context,
	payloadSource string,
	expectedType config.CollectorDataType,
) bool {
	tokenSource := ctx.GetString(collectorSourceKey)
	tokenType := ctx.GetString(collectorTypeKey)

	if tokenType != string(expectedType) {
		slog.Warn("Ingest request with invalid type", "expected", expectedType, "got", tokenType)
		respondWithUnauthorizedProblem(ctx, "invalid collector type")

		return false
	}

	if payloadSource != tokenSource {
		slog.Warn("Ingest request with source mismatch", "token_source", tokenSource, "payload_source", payloadSource)
		respondWithUnauthorizedProblem(ctx, "source mismatch: you can only push data for your own source")

		return false
	}

	return true
}
