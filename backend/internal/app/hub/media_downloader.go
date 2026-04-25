package hub

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/media"
	"github.com/potibm/billedapparat/internal/app/repository"
)


type MediaDownloader struct { 
	slideRepo repository.SlideRepository 
 }

 func NewMediaDownloader(slideRepo repository.SlideRepository) *MediaDownloader {
	return &MediaDownloader{
		slideRepo: slideRepo,
	}
}

 const (
	logPrefix = "[MediaDownloader] "
 )

func (m *MediaDownloader) ProcessSlideMedia(slideID int64) {
	ctx := context.Background()

	// 1. Load slide from DB
	slide, err := m.slideRepo.GetByID(ctx, slideID)
	if err != nil {
		slog.Error(logPrefix + "Error while fetching slide", "slideID", slideID)
		return
	}

	hasErrors := false

	// 2. Avatar
	if slide.Author != nil && slide.Author.Avatar != nil && slide.Author.Avatar.LocalURL == "" {
		localURL, err := m.resolveAndDownload(ctx, slide.Author.Avatar.OriginalURL, media.TypeAvatar)
		if err != nil {
			slog.Error(logPrefix + "Error while downloading avatar for slide", "slideID", slideID, "error", err)
			hasErrors = true
		} else {
			slide.Author.Avatar.LocalURL = localURL
			slide.Author.Avatar.MimeType = "image/webp"
		}
	}

	// 3. Content.Media
	if slide.Content.Media != nil && slide.Content.Media.LocalURL == "" {
		localURL, err := m.resolveAndDownload(ctx, slide.Content.Media.OriginalURL, media.TypeSlide)
		if err != nil {
			slog.Error(logPrefix + "Error while downloading media for slide", "slideID", slideID, "error", err)
			hasErrors = true
		} else {
			slide.Content.Media.LocalURL = localURL
			slide.Content.Media.MimeType = "image/webp"
		}
	}

	// 4. Change status
	if !hasErrors && slide.Status == domain.StatusPending {
		slide.Status = domain.StatusActive
	}

	// 5. Persist
	if err := m.slideRepo.Save(ctx, slide); err != nil {
		slog.Error(logPrefix + "Error while saving", "slideID", slideID, "error", err)
	} else {
		slog.Info(logPrefix + "Successfully processed slide", "slideID", slideID, "hasErrors", hasErrors)
	}
}

func (m *MediaDownloader) resolveAndDownload(ctx context.Context, originalURL string, imageType media.ImageType) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("original_url is empty")
	}

	// 1. Duplicate check
	existingLocalURL, found := m.slideRepo.FindLocalURLByOriginalURL(ctx, originalURL)
	if found && existingLocalURL != "" {
		slog.Debug("Cache Hit: skipping download", "originalURL", originalURL)
		return existingLocalURL, nil
	}

	// 2. when not found -> download, convert and save
	slog.Debug("Cache Miss: Download", "originalURL", originalURL)
	return m.downloadAndConvert(originalURL, imageType)
}

func (m *MediaDownloader) downloadAndConvert(originalURL string, imageType media.ImageType) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(originalURL)
	if err != nil {
		return "", fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	publicURL, err := media.ProcessAndSaveSlide(resp.Body)
	if err != nil {
		return "", fmt.Errorf("media conversion failed: %w", err)
	}

	return publicURL, nil
}