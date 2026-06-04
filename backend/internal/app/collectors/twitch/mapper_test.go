package twitch

import (
	"testing"
	"time"

	twitch "github.com/gempir/go-twitch-irc/v4"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/stretchr/testify/assert"
)

func TestMapToIngestRequest(t *testing.T) {
	fixedTime := time.Date(2026, time.May, 24, 10, 0, 0, 0, time.UTC)

	t.Run("maps message with all fields populated", func(t *testing.T) {
		message := &twitch.PrivateMessage{
			User: twitch.User{
				ID:          "user-1",
				Name:        "testuser",
				DisplayName: "TestUser",
			},
			Message: "Hello Twitch",
			ID:      "msg-123",
			Time:    fixedTime,
		}

		req := mapToIngestRequest(message, "https://example.com/avatar.png")

		assert.Equal(t, collectorName, req.Source)
		assert.Equal(t, "msg-123", req.ExternalID)
		assert.Equal(t, "Hello Twitch", req.Body)
		assert.Equal(t, "en", req.Language)
		assert.Equal(t, fixedTime, req.OriginCreatedAt)
		requireNotNilAndEqual(
			t,
			req.Author,
			"user-1",
			"testuser",
			"TestUser",
			"https://example.com/avatar.png",
		)
	})

	t.Run("maps message with empty avatar URL", func(t *testing.T) {
		message := &twitch.PrivateMessage{
			User: twitch.User{
				ID:          "user-2",
				Name:        "nouser",
				DisplayName: "NoUser",
			},
			Message: "No avatar",
			ID:      "msg-456",
			Time:    fixedTime,
		}

		req := mapToIngestRequest(message, "")

		requireNotNilAndEqual(t, req.Author, "user-2", "nouser", "NoUser", "")
	})

	t.Run("maps empty message body", func(t *testing.T) {
		message := &twitch.PrivateMessage{
			User: twitch.User{
				ID:          "user-3",
				Name:        "emptyuser",
				DisplayName: "EmptyUser",
			},
			Message: "",
			ID:      "msg-empty",
			Time:    fixedTime,
		}

		req := mapToIngestRequest(message, "")

		assert.Equal(t, "", req.Body)
		assert.Equal(t, "msg-empty", req.ExternalID)
	})
}

func requireNotNilAndEqual(
	t *testing.T,
	author *contracts.IngestSlideRequestAuthor,
	externalID, username, displayName, avatarURL string,
) {
	t.Helper()

	if author == nil {
		t.Fatal("author is nil")
	}

	assert.Equal(t, externalID, author.ExternalID)
	assert.Equal(t, username, author.Username)
	assert.Equal(t, displayName, author.DisplayName)
	assert.Equal(t, avatarURL, author.AvatarExternalURL)
}
