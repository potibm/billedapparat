package hub

import (
	"context"
	"fmt"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
)

func (s *Server) StartCollectorTextGarbageCollector(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runCollectorTextGarbageCollectionCycle()
			}
		}
	}()
}

func (s *Server) runCollectorTextGarbageCollectionCycle() {
	cutoff := time.Now().UTC().Add(-15 * time.Minute)
	ctx := context.Background()
	logger := s.logger.With("cycle", "collector_text_gc", "cutoff", cutoff)

	// 1. Find expired slides of type "social.text"
	slides, err := s.slideRepo.FindExpiredSlidesByType(ctx, "social.text", cutoff)
	if err != nil {
		logger.Error("Error finding expired slides", "type", "social.text", "error", err)

		return
	}

	logger.Debug("Found expired social text slides", "count", len(slides))

	// 2. Delete and signal frontend for each expired slide
	for _, slide := range slides {
		// Really delete the slide from the database
		err := s.slideRepo.Delete(ctx, slide.ID)
		if err != nil {
			logger.Error("Unable to delete expired slide", "id", slide.ID, "error", err)

			continue
		}

		logger.Debug("Expired social text slide cleaned up", "id", slide.ID)

		// 3. Signal to frontend that the slide has been deleted
		payload := fmt.Sprintf("%d", slide.ID)
		s.streamer.Broadcast(domain.EventDelete, []byte(payload))
	}
}
