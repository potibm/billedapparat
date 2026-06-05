package bluesky

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMapEventToIngestSlide(t *testing.T) {
	did := "did:plc:123"
	profile := &ProfileResponse{
		Handle:      "user.bsky.social",
		DisplayName: "User Display Name",
		Avatar:      "https://example.com/avatar.jpg",
	}

	t.Run("nil event commit, record or profile", func(t *testing.T) {
		assert.Nil(t, mapEventToIngestSlide(&JetstreamEvent{}, did, profile))
		assert.Nil(t, mapEventToIngestSlide(&JetstreamEvent{Commit: &JetstreamCommit{}}, did, profile))
		event := &JetstreamEvent{
			Commit: &JetstreamCommit{
				Record: &PostRecord{},
			},
		}
		assert.Nil(t, mapEventToIngestSlide(event, did, nil))
	})

	t.Run("valid event with images", func(t *testing.T) {
		createdAtStr := "2026-06-02T06:49:14Z"
		expectedTime, err := time.Parse(time.RFC3339, createdAtStr)
		assert.NoError(t, err)

		event := &JetstreamEvent{
			Did: did,
			Commit: &JetstreamCommit{
				Rkey: "rkey123",
				Record: &PostRecord{
					Text:      "Hello world #demoscene",
					CreatedAt: createdAtStr,
					Langs:     []string{"en"},
					Embed: &PostEmbed{
						Type: "app.bsky.embed.images",
						Images: []EmbedImage{
							{
								Image: MediaBlob{
									Ref: ImageRef{Link: "imgcid1"},
								},
							},
						},
					},
				},
			},
		}

		req := mapEventToIngestSlide(event, did, profile)
		assert.NotNil(t, req)
		assert.Equal(t, collectorName, req.Source)
		assert.Equal(t, "rkey123", req.ExternalID)
		assert.Equal(t, "Hello world #demoscene", req.Body)
		assert.Equal(t, "en", req.Language)
		assert.Equal(t, expectedTime, req.OriginCreatedAt)

		// Check Author
		assert.NotNil(t, req.Author)
		assert.Equal(t, did, req.Author.ExternalID)
		assert.Equal(t, profile.Handle, req.Author.Username)
		assert.Equal(t, profile.DisplayName, req.Author.DisplayName)
		assert.Equal(t, profile.Avatar, req.Author.AvatarExternalURL)

		// Check Media URLs
		assert.Len(t, req.MediaURLs, 1)
		assert.Equal(
			t,
			"https://cdn.bsky.app/img/feed_fullsize/plain/did:plc:123/imgcid1@jpeg",
			req.MediaURLs[0].ExternalURL,
		)
		assert.Equal(t, "image/jpeg", req.MediaURLs[0].ContentType)
	})

	t.Run("valid event with video embed", func(t *testing.T) {
		event := &JetstreamEvent{
			Did: did,
			Commit: &JetstreamCommit{
				Rkey: "rkey123",
				Record: &PostRecord{
					Text:      "Video post",
					CreatedAt: "invalid-time",
					Embed: &PostEmbed{
						Type: "app.bsky.embed.video",
						Video: &MediaBlob{
							Ref: ImageRef{Link: "videocid"},
						},
					},
				},
			},
		}

		req := mapEventToIngestSlide(event, did, profile)
		assert.NotNil(t, req)
		// Check fallback time
		assert.WithinDuration(t, time.Now(), req.OriginCreatedAt, 5*time.Second)

		assert.Len(t, req.MediaURLs, 1)
		assert.Equal(t, "https://video.bsky.app/watch/did:plc:123/videocid/playlist.m3u8", req.MediaURLs[0].ExternalURL)
		assert.Equal(t, "application/x-mpegURL", req.MediaURLs[0].ContentType)
	})
}
