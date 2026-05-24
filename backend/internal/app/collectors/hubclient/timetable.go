package hubclient

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendTimetableEvent(ctx context.Context, payload contracts.IngestTimetableEventRequest) error {
	return c.sendPostRequest(ctx, "timetable", "timetable", payload.ExternalID, payload)
}

func (c *HubClient) DeleteTimetableEvent(ctx context.Context, source, externalID string) error {
	return c.sendDeleteRequest(ctx, "timetable", source, externalID)
}

func (c *HubClient) SendTimetableEventSync(ctx context.Context, payload contracts.IngestTimetableSyncRequest) error {
	return c.sendPutRequest(ctx, "timetable", "timetable", payload)
}
