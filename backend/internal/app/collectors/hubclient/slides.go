package hubclient

import (
	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendSlide(payload contracts.IngestSlideRequest) error {
	return c.sendPostRequest("slide", "slide", payload.ExternalID, payload)
}

func (c *HubClient) DeleteSlide(source, externalID string) error {
	return c.sendDeleteRequest("slide", source, externalID)
}
