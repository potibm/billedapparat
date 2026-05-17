package hubclient

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendSlide(ctx context.Context, payload contracts.IngestSlideRequest) error {
	return c.sendPostRequest(ctx, "slides", "slide", payload.ExternalID, payload)
}

func (c *HubClient) DeleteSlide(ctx context.Context, source, externalID string) error {
	return c.sendDeleteRequest(ctx, "slides", source, externalID)
}
