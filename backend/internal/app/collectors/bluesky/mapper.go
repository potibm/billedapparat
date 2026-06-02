package bluesky

import (
	"time"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func mapEventToIngestSlide(event *JetstreamEvent, did string, profile *ProfileResponse) *contracts.IngestSlideRequest {
	if event.Commit == nil || event.Commit.Record == nil || profile == nil {
		return nil
	}

	author := contracts.IngestSlideRequestAuthor{
		ExternalID:        did,
		Username:          profile.Handle,
		DisplayName:       profile.DisplayName,
		AvatarExternalURL: profile.Avatar,
	}

	createdAt, err := time.Parse(time.RFC3339, event.Commit.Record.CreatedAt)
	if err != nil {
		createdAt = time.Now() // fallback to current time if parsing fails
	}

	mediaURLs := []contracts.IngestSlideRequestMediaURL{}

	images := event.ExtractImageURLs()
	for _, imageURL := range images {
		mediaURLs = append(mediaURLs, contracts.IngestSlideRequestMediaURL{
			ExternalURL: imageURL,
			ContentType: "image/jpeg",
		})
	}

	streamURL, _, hasVideo := event.ExtractVideoURLs()
	if hasVideo {
		mediaURLs = append(mediaURLs, contracts.IngestSlideRequestMediaURL{
			ExternalURL: streamURL,
			ContentType: "application/x-mpegURL",
		})
	}

	return &contracts.IngestSlideRequest{
		Source:          blueskyCollectorName,
		ExternalID:      event.Commit.Rkey,
		Author:          &author,
		Body:            event.Commit.Record.Text,
		Language:        event.Commit.Record.FirstLanguage(),
		MediaURLs:       mediaURLs,
		OriginCreatedAt: createdAt,
	}
}
