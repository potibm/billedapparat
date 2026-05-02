package hub

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/billedapparat/internal/app/domain"
)

func (s *Server) collectorIngestSlide(ctx *gin.Context) {
	var req contracts.IngestRequest
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

	// 1. Blacklist Check (Soft-Filter)
	initialStatus := s.evaluateModerationRules(ctx.Request.Context(), req)

	if initialStatus == domain.StatusFiltered {
		logFields := []any{
			"source", req.Source,
			"language", req.Language,
		}

		if req.Author != nil {
			logFields = append(logFields,
				"display_name", req.Author.DisplayName,
				"username", req.Author.Username,
			)
		} else {
			logFields = append(logFields, "author", "unknown")
		}

		slog.Info("Ingest request filtered by moderation rules", logFields...)
	}

	var createdSlideIDs []int64

	mediaPosList := make([]int, len(req.MediaURLs))
	for i := range req.MediaURLs {
		mediaPosList[i] = i
	}

	if len(req.MediaURLs) == 0 {
		mediaPosList = []int{-1}
	}

	for _, mediaPos := range mediaPosList {
		slide := mapIngestToDomain(req, mediaPos)

		exists, _ := s.slideRepo.SlideExists(slide.Source, slide.ExternalID, slide.ExternalSubID)
		if exists {
			continue
		}

		slide.Status = initialStatus

		if err := s.slideRepo.Save(ctx, &slide); err == nil {
			createdSlideIDs = append(createdSlideIDs, slide.ID)

			s.streamer.Broadcast(EventCreate, slide)
		}
	}

	if len(createdSlideIDs) == 0 {
		slog.Warn(
			"Ingest request skipped - all slides are duplicates",
			"source",
			req.Source,
			"external_id",
			req.ExternalID,
		)
		ctx.JSON(http.StatusOK, gin.H{"status": "skipped", "reason": "all duplicates"})

		return
	}

	for _, id := range createdSlideIDs {
		go s.mediaDownloader.ProcessSlideMedia(id)
	}

	slog.Info(
		"Ingest request processed",
		"source",
		req.Source,
		"external_id",
		req.ExternalID,
		"created_slides",
		len(createdSlideIDs),
	)
	ctx.JSON(http.StatusCreated, gin.H{"status": "ingested", "processed_slides": len(createdSlideIDs)})
}

func (s *Server) evaluateModerationRules(ctx context.Context, req contracts.IngestRequest) domain.SlideStatus {
	const maxRules = 1000

	rules, _, err := s.filterRuleRepo.List(ctx, maxRules, 0)
	if err != nil {
		s.logger.Error("Error fetching filter rules", "error", err)

		return domain.StatusPending
	}

	username := ""
	displayName := ""

	if req.Author != nil {
		username = req.Author.Username
		displayName = req.Author.DisplayName
	}

	for _, rule := range rules {
		if rule.Matches(req.Source, username, displayName, req.Language) {
			return domain.StatusFiltered
		}
	}

	return domain.StatusPending
}

func (s *Server) collectorDeleteSlide(ctx *gin.Context) {
	source := ctx.Param("source")
	externalID := ctx.Param("external_id")

	if source == "" || externalID == "" {
		respondWithBadRequestProblem(ctx, "provide both source and external_id in the URL")

		return
	}

	collectorToken := ctx.GetString(collectorSourceKey)
	if source != collectorToken {
		slog.Warn("Delete request with invalid source token", "source", source, "collector_token", collectorToken)
		respondWithUnauthorizedProblem(ctx, "invalid source token")

		return
	}

	err := s.slideRepo.MarkAsDeleted(ctx.Request.Context(), source, externalID)
	if err != nil {
		s.logger.Error("Error marking slide as deleted",
			"error", err,
			"source", source,
			"external_id", externalID,
		)
		respondWithInternalServerProblem(ctx, "Failed to delete slide: "+err.Error())

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "deleted",
		"source":  source,
		"id":      externalID,
	})
}
