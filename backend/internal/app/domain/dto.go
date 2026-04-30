package domain

import "time"

type (
	SlideType   string
	SlideStatus string
)

const (
	TypeSocialMedia SlideType = "social.media"
	TypeSocialText  SlideType = "social.text"
	TypeSponsor     SlideType = "sponsor"
	TypeNews        SlideType = "news"
	TypeTimetable   SlideType = "timetable"

	StatusActive   SlideStatus = "active"
	StatusPending  SlideStatus = "pending"
	StatusInactive SlideStatus = "inactive"
)

type Media struct {
	OriginalURL string `json:"original_url"`
	LocalURL    string `json:"local_url,omitempty"`
	MimeType    string `json:"mime_type"`
}

type Slide struct {
	ID              int64          `json:"id"`
	Source          string         `json:"source"`
	ExternalID      string         `json:"external_id"`
	ExternalSubID   *int           `json:"external_sub_id,omitempty"`
	Author          *Author        `json:"author"`
	Content         Content        `json:"content"`
	DisplayOptions  DisplayOptions `json:"display_options"`
	Status          SlideStatus    `json:"status"`
	OriginCreatedAt time.Time      `json:"origin_created_at"`
	CreatedAt       time.Time      `json:"created_at"`
}

type Author struct {
	ExternalID  string `json:"external_id"`
	DisplayName string `json:"display_name"`
	Avatar      *Media `json:"avatar,omitempty"`
}

type Content struct {
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Language string    `json:"language"`
	Type     SlideType `json:"type"`
	Media    *Media    `json:"media,omitempty"`
}

type DisplayOptions struct {
	AllowSocialOverlay bool `json:"allow_social_overlay"`
	Priority           int  `json:"priority"`
	IsUrgent           bool `json:"is_urgent"`
}
