package gorm

import (
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type dbSlide struct {
	GormModel

	// status and type
	Type   string `gorm:"index"` // sponsor, social, news
	Status string `gorm:"index"` // active, pending, hidden

	Source     *string `gorm:"uniqueIndex:idx_ext"`
	ExternalID *string `gorm:"uniqueIndex:idx_ext"`

	AuthorDisplayName string
	AuthorHandle      string
	AuthorAvatarURL   string

	// Content
	ContentTitle 	 string
	ContentBody      string
	MediaURLOriginal string
	MediaURLLocal    string

	// Display options
	AllowSocialOverlay bool
	Priority           int
	IsUrgent           bool

	// Metadata
	OriginCreatedAt time.Time
}

func (dbSlide) TableName() string {
	return "slides"
}

func fromDomain(s *domain.Slide) *dbSlide {
	db := &dbSlide{
		GormModel: GormModel{ID: s.ID},

		Type:               string(s.Content.Type),
		Status:             s.Status,
		AuthorDisplayName:  s.Author.DisplayName,
		AuthorHandle:       s.Author.Username,
		AuthorAvatarURL:    s.Author.AvatarURL,
		ContentTitle:       s.Content.Title,
		ContentBody:        s.Content.Body,
		MediaURLOriginal:   s.MediaURLOriginal,
		OriginCreatedAt:    s.OriginCreatedAt,
		AllowSocialOverlay: s.DisplayOptions.AllowSocialOverlay,
		IsUrgent:           s.DisplayOptions.IsUrgent,
		Priority:           s.DisplayOptions.Priority,
	}

	if s.Source != "" {
		db.Source = &s.Source
	}

	if s.ExternalID != "" {
		db.ExternalID = &s.ExternalID
	}

	return db
}

func (s *dbSlide) toDomain() *domain.Slide {
	ds := &domain.Slide{
		ID:     s.ID,
		Status: s.Status,
		Author: domain.Author{
			DisplayName: s.AuthorDisplayName,
			Username:    s.AuthorHandle,
			AvatarURL:   s.AuthorAvatarURL,
		},
		Content: domain.Content{
			Type: domain.SlideType(s.Type),
			Title: s.ContentTitle,
			Body: s.ContentBody,
		},
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: s.AllowSocialOverlay,
			IsUrgent:           s.IsUrgent,
			Priority:           s.Priority,
		},

		MediaURLOriginal: s.MediaURLOriginal,
		OriginCreatedAt:  s.OriginCreatedAt,
	}

	if s.Source != nil {
		ds.Source = *s.Source
	}

	if s.ExternalID != nil {
		ds.ExternalID = *s.ExternalID
	}

	return ds
}

func toDomainSlice(dbSlides []dbSlide) []domain.Slide {
	slides := make([]domain.Slide, len(dbSlides))
	for i, s := range dbSlides {
		slides[i] = *s.toDomain()
	}

	return slides
}
