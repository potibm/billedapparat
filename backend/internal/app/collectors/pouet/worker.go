package pouet

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *Collector) processRequest(ctx context.Context, slideRequests contracts.IngestSlideRequest) {
	if err := c.hubClient.SendSlide(ctx, slideRequests); err != nil {
		c.logger.Error("Error ingesting slide to hub", "error", err, "external_id", slideRequests.ExternalID)

		return
	}

	c.logger.Info("Successfully ingested slide to hub", "external_id", slideRequests.ExternalID)
}
