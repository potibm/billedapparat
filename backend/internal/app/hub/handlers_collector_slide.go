package hub

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
)

func (s *Server) collectorIngestSlide(ctx *gin.Context) {
	var req contracts.IngestSlideRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to parse ingest request", "error", err)
		respondWithFailedToParsePayloadProblem(ctx, err)

		return
	}

	if !s.validateCollectorAccess(ctx, req.Source, config.CollectorDataTypeSlide) {
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
		slide := mapSlideIngestToDomain(req, mediaPos)

		exists, _ := s.slideRepo.SlideExists(slide.Source, slide.ExternalID, slide.ExternalSubID)
		if exists {
			continue
		}

		slide.Status = initialStatus

		if err := s.slideRepo.Save(ctx, &slide); err == nil {
			createdSlideIDs = append(createdSlideIDs, slide.ID)

			s.streamer.Broadcast(domain.EventCreate, slide)
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

	s.incrementReceivedCollectorEventsCounter(ctx, config.CollectorDataTypeSlide, req.Source, ActionUpsert)

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

func (s *Server) evaluateModerationRules(ctx context.Context, req contracts.IngestSlideRequest) domain.SlideStatus {
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

	if !s.validateCollectorAccess(ctx, source, config.CollectorDataTypeSlide) {
		return
	}

	err := s.slideRepo.MarkAsDeleted(ctx.Request.Context(), source, externalID)
	if err != nil {
		s.logger.Error("Error marking slide as deleted",
			"error", err,
			"source", source,
			"external_id", externalID,
		)
		respondWithInternalServerProblem(ctx, "Failed to delete slide")

		return
	}

	s.incrementReceivedCollectorEventsCounter(ctx, config.CollectorDataTypeSlide, source, ActionDelete)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "deleted",
		"source":  source,
		"id":      externalID,
	})
}

func (s *Server) collectorListExternalIDs(ctx *gin.Context) {
	source := ctx.Param("source")

	if source == "" {
		respondWithBadRequestProblem(ctx, "provide source in the URL")

		return
	}

	if !s.validateCollectorAccess(ctx, source, config.CollectorDataTypeSlide) {
		return
	}

	params := repository.SlideListParams{
		ListParams: s.getListParams(ctx),
	}

	filters := repository.SlideListFilters{
		Source: &source,
	}

	slides, total, err := s.slideRepo.AdminList(ctx.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(ctx, "Failed to list slides: "+err.Error())

		return
	}

	externalIDs := make([]string, 0, len(slides))
	for _, slide := range slides {
		if slide.ExternalID != "" {
			externalIDs = append(externalIDs, slide.ExternalID)
		}
	}

	ctx.Header("X-Total-Count", strconv.FormatInt(total, 10))
	ctx.JSON(http.StatusOK, externalIDs)
}
