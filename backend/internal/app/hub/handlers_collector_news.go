package hub

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (s *Server) collectorUpsertNews(ctx *gin.Context) {
	var req contracts.IngestNewsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to parse ingest request", "error", err)
		respondWithFailedToParsePayloadProblem(ctx, err)

		return
	}

	collectorToken := ctx.GetString(collectorSourceKey)
	if req.Source != collectorToken {
		slog.Warn("Ingest request with invalid source token", "source", req.Source, "collector_token", collectorToken)
		respondWithUnauthorizedProblem(ctx, "invalid source token")

		return
	}

	news, err := mapNewsIngestToDomain(req, s.markdownConverter)
	if err != nil {
		slog.Error("Failed to map ingest news to domain model", "error", err)
		respondWithInternalServerProblem(ctx, "failed to process news item")

		return
	}

	if err := s.newsRepo.Save(ctx.Request.Context(), &news); err != nil {
		slog.Error("Failed to upsert news item", "error", err)
		respondWithInternalServerProblem(ctx, "failed to save news item")

		return
	}

	// @todo generate a slide from that!
	// s.streamer.Broadcast(EventCreate, slide)

	slog.Info("News item upserted successfully", "id", news.ID, "source", news.Source)
	ctx.JSON(http.StatusOK, gin.H{"id": news.ID})
}

func (s *Server) collectorSyncNews(ctx *gin.Context) {
}

func (s *Server) collectorDeleteNews(ctx *gin.Context) {
}
