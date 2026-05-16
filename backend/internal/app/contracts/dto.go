package contracts

import "time"

type IngestSlideRequestMediaURL struct {
	ExternalURL string `json:"external_url" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

type IngestSlideRequestAuthor struct {
	ExternalID        string `json:"external_id"          binding:"required"`
	Username          string `json:"username"             binding:"required"`
	DisplayName       string `json:"display_name"         binding:"required"`
	AvatarExternalURL string `json:"avatar_url,omitempty"`
}

type IngestSlideRequest struct {
	Source          string                       `json:"source"               binding:"required"`
	ExternalID      string                       `json:"external_id"          binding:"required"`
	Author          *IngestSlideRequestAuthor    `json:"author,omitempty"`
	Body            string                       `json:"body,omitempty"`
	MediaURLs       []IngestSlideRequestMediaURL `json:"media_urls,omitempty"`
	Language        string                       `json:"language,omitempty"`
	OriginCreatedAt time.Time                    `json:"origin_created_at"    binding:"required"`
}

type IngestNewsRequest struct {
	Source      string `json:"source"                 binding:"required"`
	ExternalID  string `json:"external_id"            binding:"required"`
	Title       string `json:"title,omitempty"        binding:"required"`
	Body        string `json:"body"                   binding:"required"`
	IsUrgent    bool   `json:"is_urgent"`
	ExternalURL string `json:"external_url,omitempty"`
	IsHidden    bool   `json:"is_hidden"`
}

type IngestNewsSyncRequest struct {
	Source string              `json:"source" binding:"required"`
	Items  []IngestNewsRequest `json:"items"  binding:"required,dive"`
}

type IngestTimetableEventRequest struct {
	Source          string    `json:"source"                     binding:"required"`
	ExternalID      string    `json:"external_id"                binding:"required"`
	Title           string    `json:"title,omitempty"            binding:"required"`
	Description     string    `json:"description,omitempty"`
	StartTime       time.Time `json:"start_time"                 binding:"required"`
	EndTime         time.Time `json:"end_time"                   binding:"required"`
	LocationName    string    `json:"location,omitempty"`
	LocationAddress string    `json:"location_address,omitempty"`
	CategoryName    string    `json:"category,omitempty"`
	CategoryColor   string    `json:"category_color,omitempty"`
	IsHidden        bool      `json:"is_hidden"`
}

type IngestTimetableSyncRequest struct {
	Source string                        `json:"source" binding:"required"`
	Items  []IngestTimetableEventRequest `json:"items"  binding:"required,dive"`
}
