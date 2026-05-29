package hub

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

const (
	timetableGeneratorEngine = "timetable-generator"
)

func (s *Server) collectorUpsertTimetable(ctx *gin.Context) {
	var req contracts.IngestTimetableEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to parse ingest request", "error", err)
		respondWithFailedToParsePayloadProblem(ctx, err)

		return
	}

	if !s.validateCollectorAccess(ctx, req.Source, config.CollectorDataTypeTimetable) {
		return
	}

	timetable := mapTimetableIngestToDomain(req)

	if err := s.timetableEventRepo.Save(ctx.Request.Context(), &timetable); err != nil {
		slog.Error("Failed to upsert timetable item", "error", err)
		respondWithInternalServerProblem(ctx, "failed to save timetable item")

		return
	}

	s.generatorEngine.Trigger(timetableGeneratorEngine)

	s.incrementReceivedCollectorEventsCounter(ctx, config.CollectorDataTypeTimetable, req.Source, ActionUpsert)

	slog.Info("Timetable item upserted successfully", "id", timetable.ID, "source", timetable.Source)
	ctx.JSON(http.StatusOK, gin.H{"id": timetable.ID})
}

func (s *Server) collectorSyncTimetable(ctx *gin.Context) {
	var req contracts.IngestTimetableSyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to parse ingest sync request", "error", err)
		respondWithFailedToParsePayloadProblem(ctx, err)

		return
	}

	if !s.validateCollectorAccess(ctx, req.Source, config.CollectorDataTypeTimetable) {
		return
	}

	timetableList, err := mapTimetableIngestListToDomain(req.Source, req.Items)
	if err != nil {
		slog.Error("Failed to map ingest timetable list to domain model", "error", err)
		respondWithInternalServerProblem(ctx, "failed to process timetable items")

		return
	}

	syncResult, err := s.timetableEventRepo.Sync(ctx.Request.Context(), req.Source, timetableList)
	if err != nil {
		slog.Error("Failed to sync timetable items", "error", err)
		respondWithInternalServerProblem(ctx, "failed to sync timetable items")

		return
	}

	if len(syncResult.Created) > 0 || len(syncResult.Updated) > 0 || len(syncResult.Deleted) > 0 {
		s.generatorEngine.Trigger(timetableGeneratorEngine)
	}

	s.incrementReceivedCollectorEventsCounter(ctx, config.CollectorDataTypeTimetable, req.Source, ActionSync)

	slog.Info(
		"Finished timetable sync",
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

func (s *Server) collectorDeleteTimetable(ctx *gin.Context) {
	source := ctx.Param("source")
	externalID := ctx.Param("external_id")

	if source == "" || externalID == "" {
		respondWithBadRequestProblem(ctx, "provide both source and external_id in the URL")

		return
	}

	if !s.validateCollectorAccess(ctx, source, config.CollectorDataTypeTimetable) {
		return
	}

	err := s.timetableEventRepo.Delete(ctx.Request.Context(), source, externalID)
	if err != nil {
		slog.Error("Error marking timetable item as deleted",
			"error", err,
			"source", source,
			"external_id", externalID,
		)
		respondWithInternalServerProblem(ctx, "Failed to delete timetable item")

		return
	}

	s.generatorEngine.Trigger(timetableGeneratorEngine)

	s.incrementReceivedCollectorEventsCounter(ctx, config.CollectorDataTypeTimetable, source, ActionDelete)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "deleted",
		"source":  source,
		"id":      externalID,
	})
}
