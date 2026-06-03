package discord

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *Collector) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.drainBuffer(ctx)

			return
		case req, ok := <-c.msgBuffer:
			if !ok {
				return
			}

			c.processRequest(ctx, req)
		}
	}
}

func (c *Collector) drainBuffer(parentCtx context.Context) {
	drainCtx := context.WithoutCancel(parentCtx)

	for len(c.msgBuffer) > 0 {
		req := <-c.msgBuffer
		c.processRequest(drainCtx, req)
	}
}

func (c *Collector) processRequest(ctx context.Context, req contracts.IngestSlideRequest) {
	if err := c.hubClient.SendSlide(ctx, req); err != nil {
		c.logger.Error("Failed to ingest slide", "error", err)
	}
}
