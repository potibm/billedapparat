package hub

import (
	"testing"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int { return &v }

func TestSmartTruncate(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		limit int
		want  string
	}{
		{
			name:  "shorter than limit",
			text:  "hello",
			limit: 10,
			want:  "hello",
		},
		{
			name:  "exactly at limit",
			text:  "hello world",
			limit: 11,
			want:  "hello world",
		},
		{
			name:  "over limit with space before limit",
			text:  "hello wonderful world",
			limit: 12,
			want:  "hello...",
		},
		{
			name:  "over limit with no space",
			text:  "supercalifragilisticexpialidocious",
			limit: 10,
			want:  "supercalif",
		},
		{
			name:  "unicode/multibyte characters",
			text:  "café ☕ morning 🌅 and more",
			limit: 13,
			want:  "café ☕...",
		},
		{
			name:  "empty string",
			text:  "",
			limit: 5,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := smartTruncate(tt.text, tt.limit)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapSlideIngestToDomain(t *testing.T) {
	now := time.Date(2026, time.May, 24, 10, 0, 0, 0, time.UTC)

	t.Run("with media URLs and mediaPos=0", func(t *testing.T) {
		author := &contracts.IngestSlideRequestAuthor{
			ExternalID:  "auth-1",
			Username:    "testuser",
			DisplayName: "Test User",
		}
		req := contracts.IngestSlideRequest{
			Source:     "test-source",
			ExternalID: "ext-123",
			Body:       "Hello world from mastodon",
			Author:     author,
			MediaURLs: []contracts.IngestSlideRequestMediaURL{
				{ExternalURL: "https://example.com/img.jpg", ContentType: "image/jpeg"},
			},
			Language:        "en",
			OriginCreatedAt: now,
		}

		slide := mapSlideIngestToDomain(req, 0)

		assert.Equal(t, domain.TypeSocialMedia, slide.Content.Type)
		assert.Equal(t, intPtr(1), slide.ExternalSubID)
		require.NotNil(t, slide.Content.Media)
		assert.Equal(t, "https://example.com/img.jpg", slide.Content.Media.OriginalURL)
		assert.Equal(t, "image/jpeg", slide.Content.Media.MimeType)
		assert.NotNil(t, slide.Author)
		assert.Equal(t, "testuser", slide.Author.Username)
		assert.Equal(t, "Hello world from mastodon", slide.Content.Body)
		assert.Equal(t, domain.StatusPending, slide.Status)
	})

	t.Run("without media URLs and mediaPos=-1", func(t *testing.T) {
		req := contracts.IngestSlideRequest{
			Source:          "test-source",
			ExternalID:      "ext-123",
			Body:            "Just a text post",
			Language:        "en",
			OriginCreatedAt: now,
		}

		slide := mapSlideIngestToDomain(req, -1)

		assert.Equal(t, domain.TypeSocialText, slide.Content.Type)
		assert.Nil(t, slide.ExternalSubID)
		assert.Nil(t, slide.Content.Media)
	})

	t.Run("with author and avatar", func(t *testing.T) {
		author := &contracts.IngestSlideRequestAuthor{
			ExternalID:        "auth-1",
			Username:          "testuser",
			DisplayName:       "Test User",
			AvatarExternalURL: "https://example.com/avatar.jpg",
		}
		req := contracts.IngestSlideRequest{
			Source:          "test-source",
			ExternalID:      "ext-123",
			Body:            "Text",
			Author:          author,
			OriginCreatedAt: now,
		}

		slide := mapSlideIngestToDomain(req, -1)

		require.NotNil(t, slide.Author)
		assert.Equal(t, "testuser", slide.Author.Username)
		require.NotNil(t, slide.Author.Avatar)
		assert.Equal(t, "https://example.com/avatar.jpg", slide.Author.Avatar.OriginalURL)
		assert.Equal(t, "image/jpeg", slide.Author.Avatar.MimeType)
	})

	t.Run("without author", func(t *testing.T) {
		req := contracts.IngestSlideRequest{
			Source:          "test-source",
			ExternalID:      "ext-123",
			Body:            "Text",
			OriginCreatedAt: now,
		}

		slide := mapSlideIngestToDomain(req, -1)

		assert.Nil(t, slide.Author)
	})

	t.Run("body used for title truncated and body", func(t *testing.T) {
		req := contracts.IngestSlideRequest{
			Source:          "test-source",
			ExternalID:      "ext-123",
			Body:            "This is a very long body text that should be truncated for title field",
			OriginCreatedAt: now,
		}

		slide := mapSlideIngestToDomain(req, -1)

		fullBody := "This is a very long body text that should be truncated for title field"
		assert.Equal(t, fullBody, slide.Content.Body)
		// Title is truncated to maxLengthTitle (30 chars via smartTruncate),
		// with "..." suffix added, so final length may exceed 30 slightly
		assert.Contains(t, slide.Content.Title, "This")
		assert.True(t, len(slide.Content.Title) < len(fullBody),
			"truncated title should be shorter than body")
	})
}

func TestMapTimetableIngestToDomain(t *testing.T) {
	startTime := time.Date(2026, time.May, 24, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, time.May, 24, 11, 0, 0, 0, time.UTC)

	t.Run("with location and category", func(t *testing.T) {
		req := contracts.IngestTimetableEventRequest{
			Source:          "test-source",
			ExternalID:      "evt-1",
			Title:           "Test Event",
			Description:     "A test event",
			StartTime:       startTime,
			EndTime:         endTime,
			LocationName:    "Room 101",
			LocationAddress: "123 Main St",
			CategoryName:    "Workshop",
			CategoryColor:   "#ff0000",
		}

		event := mapTimetableIngestToDomain(req)

		assert.Equal(t, "Test Event", event.Title)
		require.NotNil(t, event.Location)
		assert.Equal(t, "Room 101", event.Location.Name)
		assert.Equal(t, "123 Main St", event.Location.Address)
		require.NotNil(t, event.Category)
		assert.Equal(t, "Workshop", event.Category.Name)
		assert.Equal(t, "#ff0000", event.Category.Color)
	})

	t.Run("without location", func(t *testing.T) {
		req := contracts.IngestTimetableEventRequest{
			Source:     "test-source",
			ExternalID: "evt-1",
			Title:      "Test Event",
			StartTime:  startTime,
			EndTime:    endTime,
		}

		event := mapTimetableIngestToDomain(req)

		assert.Nil(t, event.Location)
	})

	t.Run("without category", func(t *testing.T) {
		req := contracts.IngestTimetableEventRequest{
			Source:       "test-source",
			ExternalID:   "evt-1",
			Title:        "Test Event",
			StartTime:    startTime,
			EndTime:      endTime,
			LocationName: "Room 101",
		}

		event := mapTimetableIngestToDomain(req)

		assert.Nil(t, event.Category)
	})
}

func TestMapTimetableIngestListToDomain(t *testing.T) {
	startTime := time.Date(2026, time.May, 24, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, time.May, 24, 11, 0, 0, 0, time.UTC)

	t.Run("all items match source", func(t *testing.T) {
		items := []contracts.IngestTimetableEventRequest{
			{Source: "src", ExternalID: "e1", Title: "E1", StartTime: startTime, EndTime: endTime},
			{Source: "src", ExternalID: "e2", Title: "E2", StartTime: startTime, EndTime: endTime},
		}

		events, err := mapTimetableIngestListToDomain("src", items)

		require.NoError(t, err)
		assert.Len(t, events, 2)
	})

	t.Run("one item mismatches source", func(t *testing.T) {
		items := []contracts.IngestTimetableEventRequest{
			{Source: "src", ExternalID: "e1", Title: "E1", StartTime: startTime, EndTime: endTime},
			{Source: "other", ExternalID: "e2", Title: "E2", StartTime: startTime, EndTime: endTime},
		}

		events, err := mapTimetableIngestListToDomain("src", items)

		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "source mismatch")
	})
}

func newTestConverter() *converter.Converter {
	return converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
		),
	)
}

func TestMapNewsIngestToDomain(t *testing.T) {
	conv := newTestConverter()

	req := contracts.IngestNewsRequest{
		Source:      "test-source",
		ExternalID:  "news-1",
		Title:       "Some News Title That Is Really Long And Should Be Truncated",
		Body:        "<p>Hello <strong>world</strong></p>",
		IsUrgent:    true,
		IsHidden:    false,
		ExternalURL: "https://example.com/news",
	}

	news, err := mapNewsIngestToDomain(req, conv)

	require.NoError(t, err)
	assert.Equal(t, "test-source", news.Source)
	assert.Equal(t, "news-1", news.ExternalID)
	assert.True(t, news.IsUrgent)
	assert.False(t, news.IsHidden)
	assert.Equal(t, "https://example.com/news", news.ExternalURL)
	// Body should be converted from HTML to markdown
	assert.Contains(t, news.Body, "world")
}

func TestMapNewsIngestListToDomain(t *testing.T) {
	conv := newTestConverter()

	t.Run("all items match source", func(t *testing.T) {
		items := []contracts.IngestNewsRequest{
			{Source: "src", ExternalID: "n1", Title: "News 1", Body: "<p>Body 1</p>"},
			{Source: "src", ExternalID: "n2", Title: "News 2", Body: "<p>Body 2</p>"},
		}

		newsList, err := mapNewsIngestListToDomain("src", items, conv)

		require.NoError(t, err)
		assert.Len(t, newsList, 2)
	})

	t.Run("one item mismatches source", func(t *testing.T) {
		items := []contracts.IngestNewsRequest{
			{Source: "src", ExternalID: "n1", Title: "News 1", Body: "<p>Body 1</p>"},
			{Source: "other", ExternalID: "n2", Title: "News 2", Body: "<p>Body 2</p>"},
		}

		newsList, err := mapNewsIngestListToDomain("src", items, conv)

		assert.Error(t, err)
		assert.Nil(t, newsList)
		assert.Contains(t, err.Error(), "source mismatch")
	})
}
