package hub

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

const (
	newsGeneratorEngine = "news-generator"
)

func (s *Server) collectorUpsertNews(ctx *gin.Context) {
	var req contracts.IngestNewsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to parse ingest request", "error", err)
		respondWithFailedToParsePayloadProblem(ctx, err)

		return
	}

	if !s.validateCollectorAccess(ctx, req.Source, config.CollectorDataTypeNews) {
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

	s.generatorEngine.Trigger(newsGeneratorEngine)

	slog.Info("News item upserted successfully", "id", news.ID, "source", news.Source)
	ctx.JSON(http.StatusOK, gin.H{"id": news.ID})
}

func (s *Server) collectorSyncNews(ctx *gin.Context) {
	var req contracts.IngestNewsSyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to parse ingest sync request", "error", err)
		respondWithFailedToParsePayloadProblem(ctx, err)

		return
	}

	if !s.validateCollectorAccess(ctx, req.Source, config.CollectorDataTypeNews) {
		return
	}

	newsList, err := mapNewsIngestListToDomain(req.Source, req.Items, s.markdownConverter)
	if err != nil {
		slog.Error("Failed to map ingest news list to domain model", "error", err)
		respondWithInternalServerProblem(ctx, "failed to process news items")

		return
	}

	syncResult, err := s.newsRepo.Sync(ctx.Request.Context(), req.Source, newsList)
	if err != nil {
		slog.Error("Failed to sync news items", "error", err)
		respondWithInternalServerProblem(ctx, "failed to sync news items")

		return
	}

	if len(syncResult.Created) > 0 || len(syncResult.Updated) > 0 || len(syncResult.Deleted) > 0 {
		s.generatorEngine.Trigger(newsGeneratorEngine)
	}

	slog.Info(
		"Finished news sync",
		"source",
		req.Source,
		"created_count",
		len(syncResult.Created),
		"updated_count",
		len(syncResult.Updated),
		"deleted_count",
		len(syncResult.Deleted),
	)
	ctx.JSON(http.StatusOK, gin.H{"message": "done"})
}

func (s *Server) collectorDeleteNews(ctx *gin.Context) {
	source := ctx.Param("source")
	externalID := ctx.Param("external_id")

	if source == "" || externalID == "" {
		respondWithBadRequestProblem(ctx, "provide both source and external_id in the URL")

		return
	}

	if !s.validateCollectorAccess(ctx, source, config.CollectorDataTypeNews) {
		return
	}

	err := s.newsRepo.Delete(ctx.Request.Context(), source, externalID)
	if err != nil {
		slog.Error("Error marking news item as deleted",
			"error", err,
			"source", source,
			"external_id", externalID,
		)
		respondWithInternalServerProblem(ctx, "Failed to delete news item: "+err.Error())

		return
	}

	s.generatorEngine.Trigger(newsGeneratorEngine)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "deleted",
		"source":  source,
		"id":      externalID,
	})
}
