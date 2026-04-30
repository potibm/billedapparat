package hub

import (
	"strings"
	"time"

	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/billedapparat/internal/app/domain"
)

func mapMediaURLToDomain(m contracts.IngestRequestMediaURL) domain.Media {
	return domain.Media{
		OriginalURL: m.ExternalURL,
		MimeType:    m.ContentType,
	}
}

func mapAuthorToDomain(a contracts.IngestRequestAuthor) domain.Author {
	avatar := (*domain.Media)(nil)
	if a.AvatarExternalURL != "" {
		avatar = &domain.Media{
			OriginalURL: a.AvatarExternalURL,
			MimeType:    "image/jpeg",
		}
	}

	return domain.Author{
		ExternalID:  a.ExternalID,
		DisplayName: a.DisplayName,
		Avatar:      avatar,
	}
}

func mapIngestToDomain(i contracts.IngestRequest, mediaPos int) domain.Slide {
	const maxLengthTitle = 30

	hasMedia := len(i.MediaURLs) > 0

	slideType := domain.TypeSocialText
	if hasMedia {
		slideType = domain.TypeSocialMedia
	}

	author := (*domain.Author)(nil)
	if i.Author != nil {
		author = new(domain.Author)
		*author = mapAuthorToDomain(*i.Author)
	}

	content := domain.Content{
		Type:     domain.SlideType(slideType),
		Title:    smartTruncate(i.Body, maxLengthTitle),
		Body:     i.Body,
		Language: i.Language,
	}

	externalSubID := (*int)(nil)
	if mediaPos >= 0 {
		externalSubID = new(int)
		*externalSubID = mediaPos + 1
	}

	slide := domain.Slide{
		Source:          i.Source,
		ExternalID:      i.ExternalID,
		ExternalSubID:   externalSubID,
		Content:         content,
		Author:          author,
		Status:          domain.StatusPending,
		OriginCreatedAt: i.OriginCreatedAt,
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: true,
			Priority:           1,
			IsUrgent:           false,
		},
		CreatedAt: time.Now(),
	}

	if mediaPos >= 0 && mediaPos < len(i.MediaURLs) {
		slide.Content.Media = new(domain.Media)
		*slide.Content.Media = mapMediaURLToDomain(i.MediaURLs[mediaPos])
	}

	return slide
}

func smartTruncate(text string, limit int) string {
	runes := []rune(text)

	if len(runes) <= limit {
		return text
	}

	subString := string(runes[:limit])
	lastSpace := strings.LastIndex(subString, " ")

	if lastSpace == -1 {
		return subString
	}

	return strings.TrimSpace(subString[:lastSpace]) + "..."
}
