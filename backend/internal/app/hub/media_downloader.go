package hub

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/media"
	"github.com/potibm/billedapparat/internal/app/repository"
)

const (
	defaultTimeout   = 5 * time.Second
	defaultKeepAlive = 30 * time.Second
)

type MediaDownloader struct {
	slideRepo repository.SlideRepository
	streamer  *Streamer
	logger    *slog.Logger
	client    *http.Client
}

func NewMediaDownloader(
	slideRepo repository.SlideRepository,
	streamer *Streamer,
	logger *slog.Logger,
) *MediaDownloader {
	return &MediaDownloader{
		slideRepo: slideRepo,
		streamer:  streamer,
		logger:    logger,
		client:    newSafeHTTPClient(),
	}
}

func (m *MediaDownloader) ProcessSlideMedia(slideID int64) {
	ctx := context.Background()

	// 1. Load slide from DB
	slide, err := m.slideRepo.GetByID(ctx, slideID)
	if err != nil {
		m.logger.Error("Unable to fetch slide from db", "slide_id", slideID)

		return
	}

	hasErrors := false

	// 2. Avatar
	if slide.Author != nil && slide.Author.Avatar != nil && slide.Author.Avatar.LocalURL == "" {
		localURL, err := m.resolveAndDownload(ctx, slide.Author.Avatar.OriginalURL, media.TypeAvatar)
		if err != nil {
			m.logger.Error("Unable to download avatar for slide", "slide_id", slideID, "error", err)

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
			m.logger.Error("Unable to download media for slide", "slide_id", slideID, "error", err)

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
		m.logger.Error("Unable to save", "slide_id", slideID, "error", err)
	} else {
		m.streamer.Broadcast(domain.EventUpdate, slide)
		m.logger.Info("Successfully processed slide", "slide_id", slideID, "has_errors", hasErrors)
	}
}

func (m *MediaDownloader) resolveAndDownload(
	ctx context.Context,
	originalURL string,
	imageType media.ImageType,
) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("original_url is empty")
	}

	// 1. Duplicate check
	existingLocalURL, found := m.slideRepo.FindLocalURLByOriginalURL(ctx, originalURL)
	if found && existingLocalURL != "" {
		m.logger.Debug("Cache Hit: skipping download", "original_url", originalURL)

		return existingLocalURL, nil
	}

	// 2. when not found -> download, convert and save
	m.logger.Debug("Cache Miss: Download", "original_url", originalURL)

	return m.downloadAndConvert(originalURL, imageType)
}

func (m *MediaDownloader) downloadAndConvert(originalURL string, imageType media.ImageType) (string, error) {
	const defaultTimeout = 15 * time.Second

	if err := validateURL(originalURL); err != nil {
		return "", fmt.Errorf("url validation failed: %w", err)
	}

	resp, err := m.client.Get(originalURL)
	if err != nil {
		return "", fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	publicURL, err := media.ProcessAndSave(resp.Body, imageType)
	if err != nil {
		return "", fmt.Errorf("media conversion failed: %w", err)
	}

	return publicURL, nil
}

func validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL schema: %s", parsedURL.Scheme)
	}

	if parsedURL.User != nil && parsedURL.User.String() != "" {
		return fmt.Errorf("urls with credentials blocked")
	}

	return nil
}

func newSafeHTTPClient() *http.Client {
	safeTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}

			var safeIP net.IP

			for _, ip := range ips {
				if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
					ip.IsUnspecified() {
					return nil, fmt.Errorf("access to internal IP %s blocked (SSRF)", ip.String())
				}

				if safeIP == nil {
					safeIP = ip
				}
			}

			if safeIP == nil {
				return nil, fmt.Errorf("no resolvable, public IP found")
			}

			safeAddr := net.JoinHostPort(safeIP.String(), port)
			dialer := &net.Dialer{
				Timeout:   defaultTimeout,
				KeepAlive: defaultKeepAlive,
			}

			return dialer.DialContext(ctx, network, safeAddr)
		},
	}

	return &http.Client{
		Timeout:   defaultTimeout,
		Transport: safeTransport,
	}
}
