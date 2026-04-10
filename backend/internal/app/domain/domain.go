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
	ID               int64
	Source           string
	ExternalID       string
	MediaURLOriginal string
	Author           Author
	Content          Content
	DisplayOptions   DisplayOptions
	Status           string
	OriginCreatedAt  time.Time
	CreatedAt        time.Time
}

type Author struct {
	ExternalID  string
	Username    string
	DisplayName string
	AvatarLocal string
}

type Content struct {
	Text      string
	Language  string
	Type      SlideType
	LocalPath string
}

type DisplayOptions struct {
	AllowSocialOverlay bool
	Priority           int
	IsUrgent           bool
}
