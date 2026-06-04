package discord

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *Collector) processRequest(ctx context.Context, req contracts.IngestSlideRequest) {
	c.logger.Info("Processing message", "external_id", req.ExternalID, "author", req.Author.Username)

	if err := c.hubClient.SendSlide(ctx, req); err != nil {
		c.logger.Error("Failed to ingest slide", "error", err)
	}
}
