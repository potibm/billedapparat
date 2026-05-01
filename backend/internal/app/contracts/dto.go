package contracts

import "time"

type IngestRequestMediaURL struct {
	ExternalURL string `json:"external_url" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

type IngestRequestAuthor struct {
	ExternalID        string `json:"external_id"          binding:"required"`
	Username          string `json:"username"             binding:"required"`
	DisplayName       string `json:"display_name"         binding:"required"`
	AvatarExternalURL string `json:"avatar_url,omitempty"`
}

type IngestRequest struct {
	Source          string                  `json:"source"               binding:"required"`
	ExternalID      string                  `json:"external_id"          binding:"required"`
	Author          *IngestRequestAuthor    `json:"author,omitempty"`
	Body            string                  `json:"body,omitempty"`
	MediaURLs       []IngestRequestMediaURL `json:"media_urls,omitempty"`
	Language        string                  `json:"language,omitempty"`
	OriginCreatedAt time.Time               `json:"origin_created_at"    binding:"required"`
}
