package twitch

import (
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

func mapToIngestRequest(message *twitch.PrivateMessage, avatarURL string) contracts.IngestSlideRequest {
	req := contracts.IngestSlideRequest{
		Source:          collectorName,
		ExternalID:      message.ID,
		Body:            message.Message,
		Language:        "en",
		OriginCreatedAt: message.Time,
		Author: &contracts.IngestSlideRequestAuthor{
			ExternalID:        message.User.ID,
			Username:          message.User.Name,
			DisplayName:       message.User.DisplayName,
			AvatarExternalURL: avatarURL,
		},
	}

	return req
}
