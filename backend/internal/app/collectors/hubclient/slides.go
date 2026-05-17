package hubclient

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendSlide(ctx context.Context, payload contracts.IngestSlideRequest) error {
	return c.sendPostRequest(ctx, "slide", "slide", payload.ExternalID, payload)
}

func (c *HubClient) DeleteSlide(ctx context.Context, source, externalID string) error {
	return c.sendDeleteRequest(ctx, "slide", source, externalID)
}
