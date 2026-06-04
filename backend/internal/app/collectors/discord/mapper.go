package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

func mapToIngestRequest(message *discordgo.Message) contracts.IngestSlideRequest {
	var mediaUrls []contracts.IngestSlideRequestMediaURL
	for _, attachment := range message.Attachments {
		mediaUrls = append(mediaUrls, contracts.IngestSlideRequestMediaURL{
			ExternalURL: attachment.URL,
			ContentType: attachment.ContentType,
		})
	}

	req := contracts.IngestSlideRequest{
		Source:          collectorName,
		ExternalID:      message.ID,
		Body:            message.Content,
		Language:        "en",
		OriginCreatedAt: message.Timestamp,
		MediaURLs:       mediaUrls,
		Author: &contracts.IngestSlideRequestAuthor{
			ExternalID:        message.Author.ID,
			Username:          message.Author.Username,
			DisplayName:       message.Author.GlobalName,
			AvatarExternalURL: message.Author.AvatarURL("128"),
		},
	}

	return req
}
