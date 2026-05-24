package mastodon

import (
	"testing"
	"time"

	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/stretchr/testify/assert"
)

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "plain text unchanged",
			content: "hello world",
			want:    "hello world",
		},
		{
			name:    "p tag with double newline",
			content: "<p>hello</p>",
			want:    "hello",
		},
		{
			name:    "p tag with text after creates double newline",
			content: "<p>hello</p>after",
			want:    "hello\n\nafter",
		},
		{
			name:    "br tag becomes single newline",
			content: "hello<br>world",
			want:    "hello\nworld",
		},
		{
			name:    "br/ tag becomes single newline",
			content: "hello<br/>world",
			want:    "hello\nworld",
		},
		{
			name:    "HTML entities unescaped",
			content: "hello &amp; world",
			want:    "hello & world",
		},
		{
			name:    "script tag stripped",
			content: "<script>alert(1)</script>hello",
			want:    "hello",
		},
		{
			name:    "nested HTML stripped",
			content: "<div><span>hello</span></div>",
			want:    "hello",
		},
		{
			name:    "empty string",
			content: "",
			want:    "",
		},
		{
			name:    "multiple paragraphs",
			content: "<p>First paragraph</p><p>Second paragraph</p>",
			want:    "First paragraph\n\nSecond paragraph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTML(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapToIngestRequest(t *testing.T) {
	fixedTime := time.Date(2026, time.May, 24, 10, 0, 0, 0, time.UTC)

	t.Run("status with image media", func(t *testing.T) {
		status := MastoStatus{
			ID:        "12345",
			CreatedAt: fixedTime,
			Content:   "<p>Hello world</p>",
			Language:  "en",
			Account: struct {
				ID          string `json:"id"`
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
				Avatar      string `json:"avatar"`
			}{
				ID:          "acct-1",
				Username:    "testuser",
				DisplayName: "Test User",
				Avatar:      "https://example.com/avatar.jpg",
			},
			MediaAttachments: []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
				URL  string `json:"url"`
			}{
				{ID: "m1", Type: "image", URL: "https://example.com/img.jpg"},
			},
		}

		req := mapToIngestRequest(status)

		assert.Equal(t, mastodonCollectorName, req.Source)
		assert.Equal(t, "12345", req.ExternalID)
		assert.Equal(t, "en", req.Language)
		assert.Equal(t, fixedTime, req.OriginCreatedAt)
		requireNotNilAndEqual(t, req.Author, "testuser", "Test User", "https://example.com/avatar.jpg")
		assert.Len(t, req.MediaURLs, 1)
		assert.Equal(t, "https://example.com/img.jpg", req.MediaURLs[0].ExternalURL)
		assert.Equal(t, "image/jpeg", req.MediaURLs[0].ContentType)
	})

	t.Run("status with video media", func(t *testing.T) {
		status := MastoStatus{
			ID:        "12346",
			CreatedAt: fixedTime,
			Content:   "<p>Video post</p>",
			Account: struct {
				ID          string `json:"id"`
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
				Avatar      string `json:"avatar"`
			}{
				ID:          "acct-2",
				Username:    "videouser",
				DisplayName: "Video User",
			},
			MediaAttachments: []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
				URL  string `json:"url"`
			}{
				{ID: "m2", Type: "video", URL: "https://example.com/video.mp4"},
			},
		}

		req := mapToIngestRequest(status)

		assert.Len(t, req.MediaURLs, 1)
		assert.Equal(t, "video/mp4", req.MediaURLs[0].ContentType)
	})

	t.Run("status with gifv media", func(t *testing.T) {
		status := MastoStatus{
			ID:        "12347",
			CreatedAt: fixedTime,
			Content:   "<p>GIF post</p>",
			Account: struct {
				ID          string `json:"id"`
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
				Avatar      string `json:"avatar"`
			}{
				ID:          "acct-3",
				Username:    "gifuser",
				DisplayName: "GIF User",
			},
			MediaAttachments: []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
				URL  string `json:"url"`
			}{
				{ID: "m3", Type: "gifv", URL: "https://example.com/animation.gifv"},
			},
		}

		req := mapToIngestRequest(status)

		assert.Len(t, req.MediaURLs, 1)
		assert.Equal(t, "video/mp4", req.MediaURLs[0].ContentType)
	})

	t.Run("status without media", func(t *testing.T) {
		status := MastoStatus{
			ID:        "12348",
			CreatedAt: fixedTime,
			Content:   "<p>Text only</p>",
			Account: struct {
				ID          string `json:"id"`
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
				Avatar      string `json:"avatar"`
			}{
				ID:          "acct-4",
				Username:    "textuser",
				DisplayName: "Text User",
			},
		}

		req := mapToIngestRequest(status)

		assert.Empty(t, req.MediaURLs)
	})

	t.Run("author with empty DisplayName falls back to Username", func(t *testing.T) {
		status := MastoStatus{
			ID:        "12349",
			CreatedAt: fixedTime,
			Content:   "<p>Test</p>",
			Account: struct {
				ID          string `json:"id"`
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
				Avatar      string `json:"avatar"`
			}{
				ID:          "acct-5",
				Username:    "nouser",
				DisplayName: "",
			},
		}

		req := mapToIngestRequest(status)

		requireNotNilAndEqual(t, req.Author, "nouser", "nouser", "")
	})

	t.Run("all fields populated correctly", func(t *testing.T) {
		status := MastoStatus{
			ID:        "ext-1",
			CreatedAt: fixedTime,
			Content:   "<p>Some body content</p>",
			Language:  "de",
			Account: struct {
				ID          string `json:"id"`
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
				Avatar      string `json:"avatar"`
			}{
				ID:          "auth-ext-1",
				Username:    "fulluser",
				DisplayName: "Full User",
				Avatar:      "https://example.com/avatar.jpg",
			},
			MediaAttachments: []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
				URL  string `json:"url"`
			}{
				{ID: "m1", Type: "image", URL: "https://example.com/img1.jpg"},
				{ID: "m2", Type: "image", URL: "https://example.com/img2.jpg"},
			},
		}

		req := mapToIngestRequest(status)

		assert.Equal(t, mastodonCollectorName, req.Source)
		assert.Equal(t, "ext-1", req.ExternalID)
		assert.Contains(t, req.Body, "Some body content")
		assert.Equal(t, "de", req.Language)
		assert.Equal(t, fixedTime, req.OriginCreatedAt)
		requireNotNilAndEqual(t, req.Author, "fulluser", "Full User", "https://example.com/avatar.jpg")
		assert.Len(t, req.MediaURLs, 2)
		assert.Equal(t, "https://example.com/img1.jpg", req.MediaURLs[0].ExternalURL)
		assert.Equal(t, "https://example.com/img2.jpg", req.MediaURLs[1].ExternalURL)
	})
}

func requireNotNilAndEqual(
	t *testing.T,
	author *contracts.IngestSlideRequestAuthor,
	username, displayName, avatarURL string,
) {
	t.Helper()

	if author == nil {
		t.Fatal("author is nil")
	}

	assert.Equal(t, username, author.Username)
	assert.Equal(t, displayName, author.DisplayName)
	assert.Equal(t, avatarURL, author.AvatarExternalURL)
}
