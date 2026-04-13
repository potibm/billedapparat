package gorm

import (
	"log/slog"
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
	ContentText      string
	MediaURLOriginal string
	MediaURLLocal    string

	// Metadata
	OriginCreatedAt time.Time
}

func (dbSlide) TableName() string {
	return "slides"
}

func fromDomain(s *domain.Slide) *dbSlide {
	db := &dbSlide{
		GormModel: GormModel{ID: s.ID},

		Type:              string(s.Content.Type),
		Status:            s.Status,
		AuthorDisplayName: s.Author.DisplayName,
		AuthorHandle:      s.Author.Username,
		AuthorAvatarURL:   s.Author.AvatarURL,
		ContentText:       s.Content.Text,
		MediaURLOriginal:  s.MediaURLOriginal,
		OriginCreatedAt:   s.OriginCreatedAt,
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
			Text: s.ContentText,
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
		slog.Info("Converting dbSlide to domain.Slide in slice", "id", s.ID, "type", s.Type, "status", s.Status)
		slides[i] = *s.toDomain()
	}

	return slides
}
