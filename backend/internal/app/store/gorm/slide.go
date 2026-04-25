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

	Source        *string `gorm:"uniqueIndex:idx_ext"`
	ExternalID    *string `gorm:"uniqueIndex:idx_ext"`
	ExternalSubID *int    `gorm:"uniqueIndex:idx_ext"`

	AuthorDisplayName       string
	AuthorHandle            string
	AuthorAvatarURLLocal    string
	AuthorAvatarURLOriginal string `gorm:"index"`
	AuthorAvatarMimeType    string

	// Content
	ContentTitle            string
	ContentBody             string
	ContentMediaURLOriginal string `gorm:"index"`
	ContentMediaURLLocal    string
	ContentMediaMimeType    string

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
		ExternalSubID:      s.ExternalSubID,
		Status:             string(s.Status),
		AuthorDisplayName:  s.Author.DisplayName,
		AuthorHandle:       s.Author.Username,
		ContentTitle:       s.Content.Title,
		ContentBody:        s.Content.Body,
		OriginCreatedAt:    s.OriginCreatedAt,
		AllowSocialOverlay: s.DisplayOptions.AllowSocialOverlay,
		IsUrgent:           s.DisplayOptions.IsUrgent,
		Priority:           s.DisplayOptions.Priority,
	}

	if s.ExternalID != "" {
		db.ExternalID = &s.ExternalID
	}

	if s.Author.Avatar != nil {
		db.AuthorAvatarURLLocal = s.Author.Avatar.LocalURL
		db.AuthorAvatarURLOriginal = s.Author.Avatar.OriginalURL
		db.AuthorAvatarMimeType = s.Author.Avatar.MimeType
	}

	if s.Content.Media != nil {
		db.ContentMediaURLLocal = s.Content.Media.LocalURL
		db.ContentMediaURLOriginal = s.Content.Media.OriginalURL
		db.ContentMediaMimeType = s.Content.Media.MimeType
	}

	if s.Source != "" {
		db.Source = &s.Source
	}

	return db
}

func (s *dbSlide) toDomain() *domain.Slide {
	ds := &domain.Slide{
		ID:     s.ID,
		Status: domain.SlideStatus(s.Status),
		Content: domain.Content{
			Type:  domain.SlideType(s.Type),
			Title: s.ContentTitle,
			Body:  s.ContentBody,
		},
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: s.AllowSocialOverlay,
			IsUrgent:           s.IsUrgent,
			Priority:           s.Priority,
		},
		OriginCreatedAt: s.OriginCreatedAt,
	}

	if s.Source != nil {
		ds.Source = *s.Source
	}

	if s.ExternalID != nil {
		ds.ExternalID = *s.ExternalID
	}

	ds.Author = &domain.Author{
		DisplayName: s.AuthorDisplayName,
		Username:    s.AuthorHandle,
	}

	if s.AuthorAvatarURLLocal != "" || s.AuthorAvatarURLOriginal != "" {
		ds.Author.Avatar = &domain.Media{
			OriginalURL: s.AuthorAvatarURLOriginal,
			LocalURL:    s.AuthorAvatarURLLocal,
			MimeType:    s.AuthorAvatarMimeType,
		}
	}

	if s.ContentMediaURLOriginal != "" || s.ContentMediaURLLocal != "" {
		ds.Content.Media = &domain.Media{
			OriginalURL: s.ContentMediaURLOriginal,
			LocalURL:    s.ContentMediaURLLocal,
			MimeType:    s.ContentMediaMimeType,
		}
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
