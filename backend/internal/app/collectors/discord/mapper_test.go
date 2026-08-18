package discord

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/stretchr/testify/assert"
)

func TestMapToIngestRequest(t *testing.T) {
	fixedTime := time.Date(2026, time.May, 24, 10, 0, 0, 0, time.UTC)

	t.Run("message with all fields populated", func(t *testing.T) {
		message := &discordgo.Message{
			ID:        "msg-123",
			Content:   "Hello from Discord",
			Timestamp: fixedTime,
			Author: &discordgo.User{
				ID:         "user-1",
				Username:   "testuser",
				GlobalName: "Test User",
				Avatar:     "avatarhash",
			},
			Attachments: []*discordgo.MessageAttachment{
				{
					ID:          "att-1",
					URL:         "https://example.com/image.png",
					ContentType: "image/png",
				},
				{
					ID:          "att-2",
					URL:         "https://example.com/video.mp4",
					ContentType: "video/mp4",
				},
			},
		}

		req, skip := mapToIngestRequest(message, false)

		assert.False(t, skip)
		assert.Equal(t, collectorName, req.Source)
		assert.Equal(t, "msg-123", req.ExternalID)
		assert.Equal(t, "Hello from Discord", req.Body)
		assert.Equal(t, "en", req.Language)
		assert.Equal(t, fixedTime, req.OriginCreatedAt)

		requireNotNilAndEqual(t, req.Author, "user-1", "testuser", "Test User")
		assert.Equal(
			t,
			"https://cdn.discordapp.com/avatars/user-1/avatarhash.png?size=128",
			req.Author.AvatarExternalURL,
		)

		assert.Len(t, req.MediaURLs, 2)
		assert.Equal(t, "https://example.com/image.png", req.MediaURLs[0].ExternalURL)
		assert.Equal(t, "image/png", req.MediaURLs[0].ContentType)
		assert.Equal(t, "https://example.com/video.mp4", req.MediaURLs[1].ExternalURL)
		assert.Equal(t, "video/mp4", req.MediaURLs[1].ContentType)
	})

	t.Run("message without attachments", func(t *testing.T) {
		message := &discordgo.Message{
			ID:        "msg-456",
			Content:   "Text only",
			Timestamp: fixedTime,
			Author: &discordgo.User{
				ID:       "user-2",
				Username: "textuser",
				Avatar:   "avatarhash2",
			},
		}

		req, skip := mapToIngestRequest(message, false)

		assert.False(t, skip)
		assert.Empty(t, req.MediaURLs)
		assert.Equal(t, "Text only", req.Body)
	})

	t.Run("author without GlobalName (fallback to Username)", func(t *testing.T) {
		message := &discordgo.Message{
			ID:        "msg-789",
			Content:   "No display name",
			Timestamp: fixedTime,
			Author: &discordgo.User{
				ID:         "user-3",
				Username:   "nouser",
				GlobalName: "",
				Avatar:     "",
			},
		}

		req, skip := mapToIngestRequest(message, false)

		assert.False(t, skip)
		requireNotNilAndEqual(t, req.Author, "user-3", "nouser", "nouser")
		assert.Equal(t, "https://cdn.discordapp.com/embed/avatars/0.png?size=128", req.Author.AvatarExternalURL)
	})

	t.Run("message with empty content", func(t *testing.T) {
		message := &discordgo.Message{
			ID:        "msg-empty",
			Content:   "",
			Timestamp: fixedTime,
			Author: &discordgo.User{
				ID:       "user-4",
				Username: "emptyuser",
			},
		}

		req, skip := mapToIngestRequest(message, false)

		assert.False(t, skip)
		assert.Equal(t, "", req.Body)
		assert.Equal(t, "msg-empty", req.ExternalID)
	})

	t.Run("dropAttachments: attachments-only message is skipped", func(t *testing.T) {
		message := &discordgo.Message{
			ID:        "msg-gif",
			Content:   "",
			Timestamp: fixedTime,
			Author: &discordgo.User{
				ID:       "user-5",
				Username: "gifuser",
			},
			Attachments: []*discordgo.MessageAttachment{
				{
					ID:          "att-gif",
					URL:         "https://example.com/meme.gif",
					ContentType: "image/gif",
				},
			},
		}

		_, skip := mapToIngestRequest(message, true)

		assert.True(t, skip)
	})

	t.Run("dropAttachments: attachments + body keeps the body", func(t *testing.T) {
		message := &discordgo.Message{
			ID:        "msg-caption",
			Content:   "look at this lol",
			Timestamp: fixedTime,
			Author: &discordgo.User{
				ID:         "user-6",
				Username:   "capuser",
				GlobalName: "Cap User",
			},
			Attachments: []*discordgo.MessageAttachment{
				{
					ID:          "att-meme",
					URL:         "https://example.com/meme.png",
					ContentType: "image/png",
				},
			},
		}

		req, skip := mapToIngestRequest(message, true)

		assert.False(t, skip)
		assert.Equal(t, "look at this lol", req.Body)
		assert.Empty(t, req.MediaURLs)
		assert.Equal(t, "msg-caption", req.ExternalID)
		requireNotNilAndEqual(t, req.Author, "user-6", "capuser", "Cap User")
	})

	t.Run("dropAttachments: text-only message is unaffected", func(t *testing.T) {
		message := &discordgo.Message{
			ID:        "msg-text",
			Content:   "just chatting",
			Timestamp: fixedTime,
			Author: &discordgo.User{
				ID:       "user-7",
				Username: "textonly",
			},
		}

		req, skip := mapToIngestRequest(message, true)

		assert.False(t, skip)
		assert.Empty(t, req.MediaURLs)
		assert.Equal(t, "just chatting", req.Body)
	})
}

func requireNotNilAndEqual(
	t *testing.T,
	author *contracts.IngestSlideRequestAuthor,
	externalID, username, displayName string,
) {
	t.Helper()

	if author == nil {
		t.Fatal("author is nil")
	}

	assert.Equal(t, externalID, author.ExternalID)
	assert.Equal(t, username, author.Username)
	assert.Equal(t, displayName, author.DisplayName)
}
