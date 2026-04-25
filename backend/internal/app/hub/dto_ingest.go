package hub

import (
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
)

/*
import "time"

type SocialIngestPayload struct {
	Platform     string    `json:"platform"      binding:"required"` // z.B. "mastodon", "discord"
	ExternalID   string    `json:"external_id"   binding:"required"`
	AuthorName   string    `json:"author_name"   binding:"required"`
	AuthorHandle string    `json:"author_handle" binding:"required"`
	AvatarURL    string    `json:"avatar_url"`
	TextContent  string    `json:"text_content"`
	MediaURL     string    `json:"media_url"`
	PostedAt     time.Time `json:"posted_at"     binding:"required"`
}
*/


type IngestRequestMediaURL struct {
	ExternalURL string `json:"external_url"       binding:"required"`
	ContentType string `json:"content_type"       binding:"required"`
}

type IngestRequestAuthor struct {
	ExternalID        string `json:"external_id"       binding:"required"`
	DisplayName       string `json:"display_name"      binding:"required"`
	AvatarExternalURL string `json:"avatar_url,omitempty"`
}

type IngestRequest struct {
	Source          string                  `json:"source"            binding:"required"`
	ExternalID      string                  `json:"external_id"       binding:"required"`
	Author          *IngestRequestAuthor    `json:"author,omitempty"            `
	Body            string                  `json:"body,omitempty"              `
	MediaURLs       []IngestRequestMediaURL `json:"media_urls,omitempty"`
	Language        string                  `json:"language,omitempty"          binding:"required"`
	OriginCreatedAt time.Time               `json:"origin_created_at" binding:"required"`
}

func (m IngestRequestMediaURL) toDomain() domain.Media {
	return domain.Media{
		OriginalURL: m.ExternalURL,
		MimeType:    m.ContentType,
	}
}

func (a IngestRequestAuthor) toDomain() domain.Author {
	avatar := (*domain.Media)(nil)
	if a.AvatarExternalURL != "" {
		avatar = &domain.Media{
			OriginalURL: a.AvatarExternalURL,
			MimeType:   "image/jpeg",
		}
	}

	return domain.Author{
		ExternalID:  a.ExternalID,
		DisplayName: a.DisplayName,
		Avatar:      avatar,
	}
}

func (i IngestRequest) toDomain(mediaPos int) domain.Slide {
	author := (*domain.Author)(nil)
	if i.Author != nil {
		author = new(domain.Author)
		*author = i.Author.toDomain()
	}

	content := domain.Content{
		Type:     domain.SlideType(domain.TypeSocial),
		Title:    "",
		Body:     i.Body,
		Language: i.Language,
	}

	externalSubID := (*int)(nil)
	if mediaPos >= 0 {
		externalSubID = new(int)
		*externalSubID = mediaPos + 1
	}

	slide := domain.Slide{
		Source:          i.Source,
		ExternalID:      i.ExternalID,
		ExternalSubID:   externalSubID,
		Content: content,
		Author: author,
		Status: domain.StatusPending,
		OriginCreatedAt: i.OriginCreatedAt,
		CreatedAt:       time.Now(),
	}

	if mediaPos >= 0 && mediaPos < len(i.MediaURLs) {
		slide.Content.Media = new(domain.Media)
		*slide.Content.Media = i.MediaURLs[mediaPos].toDomain()
	}
	
	return slide
}