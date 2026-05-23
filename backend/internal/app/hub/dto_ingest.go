package hub

import (
	"fmt"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/billedapparat/internal/app/domain"
)

func mapMediaURLToDomain(m contracts.IngestSlideRequestMediaURL) domain.Media {
	return domain.Media{
		OriginalURL: m.ExternalURL,
		MimeType:    m.ContentType,
	}
}

func mapAuthorToDomain(a contracts.IngestSlideRequestAuthor) domain.Author {
	avatar := (*domain.Media)(nil)
	if a.AvatarExternalURL != "" {
		avatar = &domain.Media{
			OriginalURL: a.AvatarExternalURL,
			MimeType:    "image/jpeg",
		}
	}

	return domain.Author{
		ExternalID:  a.ExternalID,
		Username:    a.Username,
		DisplayName: a.DisplayName,
		Avatar:      avatar,
	}
}

func mapSlideIngestToDomain(i contracts.IngestSlideRequest, mediaPos int) domain.Slide {
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

func mapNewsIngestToDomain(i contracts.IngestNewsRequest, conv *converter.Converter) (domain.News, error) {
	const maxLengthTitle = 100

	body, err := conv.ConvertString(i.Body)
	if err != nil {
		return domain.News{}, fmt.Errorf("failed to convert markdown to text: %w", err)
	}

	news := domain.News{
		Source:      i.Source,
		ExternalID:  i.ExternalID,
		Title:       smartTruncate(i.Title, maxLengthTitle),
		Body:        body,
		IsUrgent:    i.IsUrgent,
		IsHidden:    i.IsHidden,
		ExternalURL: i.ExternalURL,
	}

	return news, nil
}

func mapNewsIngestListToDomain(
	source string,
	items []contracts.IngestNewsRequest,
	conv *converter.Converter,
) ([]domain.News, error) {
	var newsList []domain.News

	for _, item := range items {
		if item.Source != source {
			return nil, fmt.Errorf("source mismatch: expected %s, got %s", source, item.Source)
		}

		newsItem, err := mapNewsIngestToDomain(item, conv)
		if err != nil {
			return nil, fmt.Errorf("failed to map news ingest to domain: %w", err)
		}

		newsList = append(newsList, newsItem)
	}

	return newsList, nil
}

func mapTimetableIngestToDomain(i contracts.IngestTimetableEventRequest) domain.TimetableEvent {
	var location *domain.Location
	if i.LocationName != "" || i.LocationAddress != "" {
		location = &domain.Location{
			Name:    i.LocationName,
			Address: i.LocationAddress,
		}
	}

	var category *domain.Category
	if i.CategoryName != "" || i.CategoryColor != "" {
		category = &domain.Category{
			Name:  i.CategoryName,
			Color: i.CategoryColor,
		}
	}

	timetableEvent := domain.TimetableEvent{
		Source:      i.Source,
		ExternalID:  i.ExternalID,
		Title:       i.Title,
		Description: i.Description,
		StartTime:   i.StartTime,
		EndTime:     i.EndTime,
		IsHidden:    i.IsHidden,
		Location:    location,
		Category:    category,
	}

	return timetableEvent
}

func mapTimetableIngestListToDomain(
	source string,
	items []contracts.IngestTimetableEventRequest,
) ([]domain.TimetableEvent, error) {
	var timetableEvents []domain.TimetableEvent

	for _, item := range items {
		if item.Source != source {
			return nil, fmt.Errorf("source mismatch: expected %s, got %s", source, item.Source)
		}

		timetableEvent := mapTimetableIngestToDomain(item)
		timetableEvents = append(timetableEvents, timetableEvent)
	}

	return timetableEvents, nil
}
