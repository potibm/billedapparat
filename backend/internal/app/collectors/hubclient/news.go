package hubclient

import (
	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendNews(payload contracts.IngestNewsRequest) error {
	return c.sendPostRequest("news", "news", payload.ExternalID, payload)
}

func (c *HubClient) DeleteNews(source, externalID string) error {
	return c.sendDeleteRequest("news", source, externalID)
}

func (c *HubClient) SendNewsSync(payload contracts.IngestNewsSyncRequest) error {
	return c.sendPutRequest("news", "news", payload)
}
