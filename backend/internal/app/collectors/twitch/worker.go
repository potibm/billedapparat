package twitch

import (
	"context"

	"github.com/gempir/go-twitch-irc/v4"
)

func (c *Collector) processRequest(ctx context.Context, m twitch.PrivateMessage) {
	c.logger.Info("Processing message", "external_id", m.ID, "author", m.User.Name)

	avatarURL := ""

	if c.twitchHelixClient != nil {
		retrievedAvatarURL, err := c.getAvatarURL(ctx, m.User.ID)
		if err != nil {
			c.logger.Error("Failed to get avatar URL", "error", err)
		} else {
			c.logger.Info("Fetched avatar URL", "url", retrievedAvatarURL)
			avatarURL = retrievedAvatarURL
		}
	}

	req := mapToIngestRequest(&m, avatarURL)

	if err := c.hubClient.SendSlide(ctx, req); err != nil {
		c.logger.Error("Failed to ingest slide", "error", err)
	}
}
