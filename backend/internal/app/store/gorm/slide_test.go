package gorm

import (
	"testing"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
)

func TestSlideMapping(t *testing.T) {
	now := time.Now().Truncate(time.Second) // DBs speichern oft keine Nanosekunden
	extID := "ext_123"
	source := "mastodon"
	subID := 1

	t.Run("Full Mapping: Domain to DB and back", func(t *testing.T) {
		original := &domain.Slide{
			ID:            100,
			Source:        source,
			ExternalID:    extID,
			ExternalSubID: &subID,
			Status:        domain.StatusActive,
			Content: domain.Content{
				Type:     domain.TypeNews,
				Title:    "Headline",
				Body:     "Body Text",
				Language: "de",
				Media: &domain.Media{
					OriginalURL: "orig_media",
					LocalURL:    "local_media",
					MimeType:    "image/jpeg",
				},
			},
			Author: &domain.Author{
				Username:    "johndoe",
				DisplayName: "John",
				ExternalID:  "auth_123",
				Avatar: &domain.Media{
					OriginalURL: "orig_av",
					LocalURL:    "local_av",
					MimeType:    "image/png",
				},
			},
			DisplayOptions: domain.DisplayOptions{
				AllowSocialOverlay: true,
				Priority:           5,
				IsUrgent:           true,
			},
			OriginCreatedAt: now,
		}

		// To DB
		dbModel := fromDomainSlide(original)
		assert.Equal(t, int64(100), dbModel.ID)
		assert.Equal(t, extID, *dbModel.ExternalID)
		assert.Equal(t, source, *dbModel.Source)
		assert.Equal(t, subID, *dbModel.ExternalSubID)
		assert.Equal(t, "news", dbModel.Type)

		// Back to Domain
		back := dbModel.toDomain()

		// Deep check
		assert.Equal(t, original.ID, back.ID)
		assert.Equal(t, original.Source, back.Source)
		assert.Equal(t, original.ExternalID, back.ExternalID)
		assert.Equal(t, *original.ExternalSubID, *back.ExternalSubID)
		assert.Equal(t, original.Content.Title, back.Content.Title)
		assert.Equal(t, original.Author.Username, back.Author.Username)
		assert.Equal(t, original.Author.Avatar.LocalURL, back.Author.Avatar.LocalURL)
		assert.Equal(t, original.DisplayOptions.IsUrgent, back.DisplayOptions.IsUrgent)
		assert.True(t, original.OriginCreatedAt.Equal(back.OriginCreatedAt))
	})

	t.Run("Minimal Mapping: Nil handling", func(t *testing.T) {
		// Teste was passiert wenn Author und Media fehlen
		minDomain := &domain.Slide{
			ID:     1,
			Status: domain.StatusPending,
			Content: domain.Content{
				Type: domain.TypeSocialMedia,
			},
		}

		dbModel := fromDomainSlide(minDomain)
		assert.Nil(t, dbModel.ExternalID)
		assert.Nil(t, dbModel.Source)
		assert.Empty(t, dbModel.AuthorUsername)

		back := dbModel.toDomain()
		assert.Nil(t, back.Author)
		assert.Nil(t, back.Content.Media)
		assert.Empty(t, back.Source)
	})

	t.Run("Slice mapping", func(t *testing.T) {
		dbSlides := []dbSlide{
			{GormModel: GormModel{ID: 1}, Type: "news"},
			{GormModel: GormModel{ID: 2}, Type: "sponsor"},
		}
		res := toDomainSlideList(dbSlides)
		assert.Len(t, res, 2)
		assert.Equal(t, domain.SlideType("news"), res[0].Content.Type)
	})
}
