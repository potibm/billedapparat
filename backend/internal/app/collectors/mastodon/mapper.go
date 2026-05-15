package mastodon

import (
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

var (
	stripHTMLPolicy = bluemonday.StrictPolicy()
	reBR            = regexp.MustCompile(`(?i)<br\s*/?>`)
	reP             = regexp.MustCompile(`(?i)</p>`)
)

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
	content = reBR.ReplaceAllString(content, "\n")
	content = reP.ReplaceAllString(content, "\n\n")

	cleanBody := stripHTMLPolicy.Sanitize(content)

	cleanBody = html.UnescapeString(cleanBody)

	cleanBody = strings.TrimSpace(cleanBody)

	return cleanBody
}

func mapToIngestRequest(status MastoStatus) contracts.IngestSlideRequest {
	req := contracts.IngestSlideRequest{
		Source:          mastodonCollectorName,
		ExternalID:      status.ID,
		Body:            stripHTML(status.Content),
		Language:        status.Language,
		OriginCreatedAt: status.CreatedAt,
		Author: &contracts.IngestSlideRequestAuthor{
			ExternalID:        status.Account.ID,
			Username:          status.Account.Username,
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

		req.MediaURLs = append(req.MediaURLs, contracts.IngestSlideRequestMediaURL{
			ExternalURL: media.URL,
			ContentType: mimeType,
		})
	}

	return req
}
