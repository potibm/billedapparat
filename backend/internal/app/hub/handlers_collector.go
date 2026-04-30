package hub

import (
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	// @TODO need to check that the source matches the token

	// 1. Blacklist Check (Soft-Filter)
	initialStatus := domain.StatusPending
	/*if s.blacklist.IsBanned(req.Content.Text) || s.blacklist.IsBanned(req.Author.Username) {
		initialStatus = domain.StatusFiltered
	}*/

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
