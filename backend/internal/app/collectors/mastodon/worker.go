package mastodon

import "context"

func (c *Collector) handleEvent(ctx context.Context, event Event) {
	switch event.Type {
	case EventTypeUpdate, EventStatusUpdate:
		var status MastoStatus

		c.logger.Debug("Received new status", "payload", event.Payload)

		status, err := event.Status()
		if err != nil {
			c.logger.Error("Error parsing status", "error", err)

			return
		}

		c.logger.Info("Received new post", "id", status.ID, "author", status.Account.Username)

		req := mapToIngestRequest(status)

		if err := c.hubClient.SendSlide(ctx, req); err != nil {
			c.logger.Error("Error sending the post to the hub", "error", err)
		}

	case EventTypeDelete:
		statusID := event.Payload
		c.logger.Info("Post was deleted", "id", statusID, "event_type", event.Type)

		if err := c.hubClient.DeleteSlide(ctx, collectorName, statusID); err != nil {
			c.logger.Error("Error sending delete request to the hub", "error", err)
		}
	default:
		c.logger.Info("Received unsupported event type", "type", event.Type)
	}
}
