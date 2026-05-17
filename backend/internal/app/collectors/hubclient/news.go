package hubclient

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendNews(ctx context.Context, payload contracts.IngestNewsRequest) error {
	return c.sendPostRequest(ctx, "news", "news", payload.ExternalID, payload)
}

func (c *HubClient) DeleteNews(ctx context.Context, source, externalID string) error {
	return c.sendDeleteRequest(ctx, "news", source, externalID)
}

func (c *HubClient) SendNewsSync(ctx context.Context, payload contracts.IngestNewsSyncRequest) error {
	return c.sendPutRequest(ctx, "news", "news", payload)
}
