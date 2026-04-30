package mastodon

import (
	"html"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

var stripHTMLPolicy = bluemonday.StrictPolicy()

type MastoStatus struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	Account   struct {
		ID          string `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"account"`
	MediaAttachments []struct {
		ID   string `json:"id"`
		Type string `json:"type"` // "image", "gifv", "video"
		URL  string `json:"url"`
	} `json:"media_attachments"`
}

func stripHTML(content string) string {
	cleanBody := stripHTMLPolicy.Sanitize(content)

	cleanBody = html.UnescapeString(cleanBody)

	cleanBody = strings.TrimSpace(cleanBody)

	return cleanBody
}

func mapToIngestRequest(status MastoStatus) contracts.IngestRequest {
	req := contracts.IngestRequest{
		Source:          mastodonCollectorName,
		ExternalID:      status.ID,
		Body:            stripHTML(status.Content),
		Language:        status.Language,
		OriginCreatedAt: status.CreatedAt,
		Author: &contracts.IngestRequestAuthor{
			ExternalID:        status.Account.ID,
			DisplayName:       status.Account.DisplayName,
			AvatarExternalURL: status.Account.Avatar,
		},
	}

	if req.Author.DisplayName == "" {
		req.Author.DisplayName = status.Account.Username
	}

	for _, media := range status.MediaAttachments {
		mimeType := "image/jpeg" // Fallback
		if media.Type == "video" || media.Type == "gifv" {
			mimeType = "video/mp4"
		}

		req.MediaURLs = append(req.MediaURLs, contracts.IngestRequestMediaURL{
			ExternalURL: media.URL,
			ContentType: mimeType,
		})
	}

	return req
}
