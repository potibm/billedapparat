package hub

import "time"

// SocialIngestPayload ist alles, was der Hub von den externen Scrapern erwartet.
type SocialIngestPayload struct {
	Platform     string    `json:"platform"      binding:"required"` // z.B. "mastodon", "discord"
	ExternalID   string    `json:"external_id"   binding:"required"`
	AuthorName   string    `json:"author_name"   binding:"required"`
	AuthorHandle string    `json:"author_handle" binding:"required"`
	AvatarURL    string    `json:"avatar_url"`
	TextContent  string    `json:"text_content"`
	MediaURL     string    `json:"media_url"` // Das Original-Bild aus dem Netz
	PostedAt     time.Time `json:"posted_at"     binding:"required"`
}
