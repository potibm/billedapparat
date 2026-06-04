package twitch

import (
	"context"
	"fmt"

	"github.com/nicklaw5/helix/v2"
)

func (c *Collector) getAvatarURL(ctx context.Context, userID string) (string, error) {
	if c.twitchHelixClient == nil {
		return "", nil
	}

	if url, found := c.avatarCache.Get(userID); found {
		c.logger.Debug("Found avatar URL in cache", "url", url)

		return url, nil
	}

	userResp, err := c.twitchHelixClient.GetUsers(&helix.UsersParams{
		IDs: []string{userID},
	})
	if err != nil {
		return "", err
	}

	fmt.Printf("%+v\n", userResp)

	if len(userResp.Data.Users) == 0 {
		return "", nil
	}

	avatarURL := userResp.Data.Users[0].ProfileImageURL
	c.avatarCache.Set(userID, avatarURL)

	c.logger.Debug("Fetched avatar URL from Twitch API", "url", avatarURL)

	return avatarURL, nil
}
