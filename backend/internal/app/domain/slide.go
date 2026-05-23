package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type (
	StreamEvent string
	SlideType   string
	SlideStatus string
)

const (
	EventCreate StreamEvent = "CREATE"
	EventUpdate StreamEvent = "UPDATE"
	EventDelete StreamEvent = "DELETE"

	TypeSocialMedia SlideType = "social.media"
	TypeSocialText  SlideType = "social.text"
	TypeSponsor     SlideType = "sponsor"
	TypeNews        SlideType = "news"
	TypeTimetable   SlideType = "timetable"

	StatusActive   SlideStatus = "active"
	StatusPending  SlideStatus = "pending"
	StatusFiltered SlideStatus = "filtered"
	StatusInactive SlideStatus = "inactive"
	StatusDeleted  SlideStatus = "deleted"
)

func IsValidSlideType(t SlideType) bool {
	switch t {
	case TypeSocialMedia, TypeSocialText, TypeSponsor, TypeNews, TypeTimetable:
		return true
	default:
		return false
	}
}

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
	DeletedAt       *time.Time     `json:"deleted_at,omitempty"`
}

type Author struct {
	ExternalID  string `json:"external_id"`
	Username    string `json:"username"`
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

func (s Slide) HasChanged(newSlide Slide) bool {
	if s.Status != newSlide.Status {
		return true
	}

	type visualState struct {
		Content        Content
		DisplayOptions DisplayOptions
	}

	oldState, errOld := json.Marshal(visualState{s.Content, s.DisplayOptions})
	newState, errNew := json.Marshal(visualState{newSlide.Content, newSlide.DisplayOptions})

	if errOld != nil || errNew != nil {
		return true
	}

	return !bytes.Equal(oldState, newState)
}

func (s Slide) SyncKey() string {
	if s.ExternalSubID != nil {
		return fmt.Sprintf("%s|%d", s.ExternalID, *s.ExternalSubID)
	}

	return s.ExternalID + "|"
}
