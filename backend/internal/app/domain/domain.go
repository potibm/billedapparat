package domain

import "time"

type SlideType string

const (
	TypeSocial    SlideType = "social"
	TypeSponsor   SlideType = "sponsor"
	TypeNews      SlideType = "news"
	TypeTimetable SlideType = "timetable"
)

type Slide struct {
	ID               int64          `json:"id"`
	Source           string         `json:"source"`
	ExternalID       string         `json:"external_id"`
	MediaURLOriginal string         `json:"media_url_original"`
	Author           Author         `json:"author"`
	Content          Content        `json:"content"`
	DisplayOptions   DisplayOptions `json:"display_options"`
	Status           string         `json:"status"`
	OriginCreatedAt  time.Time      `json:"origin_created_at"`
	CreatedAt        time.Time      `json:"created_at"`
}

type Author struct {
	ExternalID  string `json:"external_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	AvatarLocal string `json:"avatar_local"`
}

type Content struct {
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Language  string    `json:"language"`
	Type      SlideType `json:"type"`
	LocalPath string    `json:"local_path"`
}

type DisplayOptions struct {
	AllowSocialOverlay bool `json:"allow_social_overlay"`
	Priority           int  `json:"priority"`
	IsUrgent           bool `json:"is_urgent"`
}
