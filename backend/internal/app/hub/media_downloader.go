package hub

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/media"
	"github.com/potibm/billedapparat/internal/app/repository"
	"golang.org/x/sync/singleflight"
	"golang.org/x/time/rate"
)

const (
	appRepoURL = "https://github.com/potibm/billedapparat"

	// defaultTimeout begrenzt den gesamten Download inklusive Body-Lesen.
	defaultTimeout = 30 * time.Second
	// dialTimeout begrenzt nur den Verbindungsaufbau.
	dialTimeout      = 5 * time.Second
	defaultKeepAlive = 30 * time.Second

	// maxParallelFetch begrenzt gleichzeitige CDN-Verbindungen.
	maxParallelFetch = 4
	// fetchRatePerSec bleibt bewusst unter Discords Global-Limit von 50 req/s.
	fetchRatePerSec = 20
	fetchBurst      = 10

	maxFetchAttempts = 3
	// maxRetryWait kappt sehr große Retry-After-Werte, damit Worker nicht blockieren.
	maxRetryWait = 60 * time.Second
)

// isRetryableStatus beschreibt echte transiente Serverfehler, die wiederholt werden
// dürfen. 401, 403 und 404 sind bewusst nicht enthalten: Discord zählt sie in das
// Invalid-Request-Budget (10.000 pro 10 Minuten), ein 403 auf dem CDN bedeutet
// abgelaufene URL-Signatur oder Cloudflare-Block und bleibt auch beim Retry bestehen.
func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

type MediaDownloader struct {
	slideRepo repository.SlideRepository
	streamer  *Streamer
	logger    *slog.Logger
	client    *http.Client
	semaphore chan struct{}
	limiter   *rate.Limiter
	version   string

	// avatarCache bildet originalURL -> publicURL ab. Discord-Avatar-URLs enthalten
	// den Content-Hash und ändern sich nur bei einem neuen Profilbild.
	avatarCache   map[string]string
	avatarCacheMu sync.RWMutex
	avatarFlight  singleflight.Group

	// retryBackoff berechnet das Warten zwischen zwei Versuchen. Als Feld gehalten,
	// damit Tests ohne echte Wartezeiten laufen können.
	retryBackoff func(attempt int) time.Duration
}

func NewMediaDownloader(
	slideRepo repository.SlideRepository,
	streamer *Streamer,
	logger *slog.Logger,
	version string,
) *MediaDownloader {
	return newMediaDownloader(slideRepo, streamer, logger, version, newSafeHTTPClient())
}

func newMediaDownloader(
	slideRepo repository.SlideRepository,
	streamer *Streamer,
	logger *slog.Logger,
	version string,
	client *http.Client,
) *MediaDownloader {
	return &MediaDownloader{
		slideRepo:    slideRepo,
		streamer:     streamer,
		logger:       logger,
		client:       client,
		semaphore:    make(chan struct{}, maxParallelFetch),
		limiter:      rate.NewLimiter(rate.Limit(fetchRatePerSec), fetchBurst),
		version:      version,
		avatarCache:  make(map[string]string),
		retryBackoff: backoffFor,
	}
}

func (m *MediaDownloader) ProcessSlideMedia(ctx context.Context, slideID int64) {
	// 1. Load slide from DB
	slide, err := m.slideRepo.GetByID(ctx, slideID)
	if err != nil {
		m.logger.Error("Unable to fetch slide from db", "slide_id", slideID)

		return
	}

	hasErrors := false

	// 2. Avatar
	if slide.Author != nil && slide.Author.Avatar != nil && slide.Author.Avatar.LocalURL == "" {
		localURL, err := m.resolveAvatarURL(ctx, slide.Author.Avatar.OriginalURL)
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

// resolveAvatarURL liefert die lokale publicURL eines Avatars aus dem Cache oder lädt
// ihn genau einmal pro URL. singleflight verhindert parallele Downloads derselben URL.
func (m *MediaDownloader) resolveAvatarURL(ctx context.Context, originalURL string) (string, error) {
	return m.cachedAvatarURL(ctx, originalURL, func(ctx context.Context, u string) (string, error) {
		return m.downloadAndConvert(ctx, u, media.TypeAvatar)
	})
}

func (m *MediaDownloader) cachedAvatarURL(
	ctx context.Context,
	originalURL string,
	download func(context.Context, string) (string, error),
) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("original_url is empty")
	}

	// 1. Cache-Hit?
	m.avatarCacheMu.RLock()
	cached, ok := m.avatarCache[originalURL]
	m.avatarCacheMu.RUnlock()

	if ok && cached != "" {
		m.logger.Debug("Avatar cache hit", "original_url", originalURL)

		return cached, nil
	}

	// 2. Nur ein Download pro URL
	result, err, _ := m.avatarFlight.Do(originalURL, func() (any, error) {
		// Double-check: eine parallele Goroutine kann den Cache inzwischen gefüllt haben.
		m.avatarCacheMu.RLock()
		cached, ok := m.avatarCache[originalURL]
		m.avatarCacheMu.RUnlock()

		if ok && cached != "" {
			return cached, nil
		}

		publicURL, err := download(ctx, originalURL)
		if err != nil {
			return "", err
		}

		m.avatarCacheMu.Lock()
		m.avatarCache[originalURL] = publicURL
		m.avatarCacheMu.Unlock()

		m.logger.Debug("Avatar cached", "original_url", originalURL, "public_url", publicURL)

		return publicURL, nil
	})
	if err != nil {
		return "", err
	}

	publicURL, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected singleflight result type %T", result)
	}

	return publicURL, nil
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

	return m.downloadAndConvert(ctx, originalURL, imageType)
}

func (m *MediaDownloader) downloadAndConvert(
	ctx context.Context,
	originalURL string,
	imageType media.ImageType,
) (string, error) {
	if err := validateURL(originalURL); err != nil {
		return "", fmt.Errorf("url validation failed: %w", err)
	}

	// 1. Parallelität begrenzen
	select {
	case m.semaphore <- struct{}{}:
		defer func() { <-m.semaphore }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// 2. Rate-Limit respektieren
	if err := m.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	resp, err := m.fetchWithRetry(ctx, originalURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	publicURL, err := media.ProcessAndSave(resp.Body, imageType)
	if err != nil {
		return "", fmt.Errorf("media conversion failed: %w", err)
	}

	return publicURL, nil
}

// retryPlan beschreibt einen wiederholbaren Fehlversuch: wie lange gewartet werden
// soll und welcher Fehler nach dem letzten Versuch gemeldet wird.
type retryPlan struct {
	wait time.Duration
	err  error
}

// fetchWithRetry lädt die URL und wiederholt ausschließlich 429 und echte 5xx.
func (m *MediaDownloader) fetchWithRetry(ctx context.Context, originalURL string) (*http.Response, error) {
	userAgent := m.userAgentForURL(originalURL)

	var lastErr error

	for attempt := 1; attempt <= maxFetchAttempts; attempt++ {
		resp, plan, err := m.doRequest(ctx, originalURL, userAgent, attempt)
		if err != nil {
			return nil, err
		}

		if resp != nil {
			return resp, nil
		}

		lastErr = plan.err

		if attempt == maxFetchAttempts {
			break
		}

		if err := sleepCtx(ctx, plan.wait); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("giving up after %d attempts: %w", maxFetchAttempts, lastErr)
}

// doRequest führt einen einzelnen Versuch aus. Bei Erfolg wird die Response
// zurückgegeben, bei einem wiederholbaren Fehler ein retryPlan. Nicht wiederholbare
// Fehler und Context-Abbrüche werden direkt als error zurückgegeben.
func (m *MediaDownloader) doRequest(
	ctx context.Context,
	originalURL, userAgent string,
	attempt int,
) (*http.Response, retryPlan, error) {
	resp, err := m.fetchOnce(ctx, originalURL, userAgent)
	if err != nil {
		// Abbruch durch Shutdown/Timeout nicht als transienten Fehler behandeln.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, retryPlan{}, ctxErr
		}

		return nil, m.retryPlanFor(attempt, err), nil
	}

	switch {
	case resp.StatusCode == http.StatusOK:
		return resp, retryPlan{}, nil

	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, m.rateLimitPlan(resp, originalURL, attempt), nil

	case isRetryableStatus(resp.StatusCode):
		resp.Body.Close()

		return nil, m.retryPlanFor(attempt, fmt.Errorf("transient server error: %s", resp.Status)), nil

	default:
		// 401, 403, 404 und alle übrigen: kein Retry.
		resp.Body.Close()

		m.logger.Error("Non-retryable HTTP error",
			"url", originalURL, "status", resp.StatusCode, "user_agent", userAgent)

		return nil, retryPlan{}, fmt.Errorf("bad status: %s", resp.Status)
	}
}

func (m *MediaDownloader) retryPlanFor(attempt int, err error) retryPlan {
	return retryPlan{wait: m.retryBackoff(attempt), err: err}
}

// rateLimitPlan baut den Retry-Plan für einen 429er inklusive Logging.
func (m *MediaDownloader) rateLimitPlan(resp *http.Response, originalURL string, attempt int) retryPlan {
	wait := retryDelay(resp, m.retryBackoff(attempt))

	m.logger.Warn("CDN rate limited, backing off",
		"url", originalURL, "retry_after", wait.String(), "attempt", attempt)

	resp.Body.Close()

	return retryPlan{wait: wait, err: fmt.Errorf("rate limited: %s", resp.Status)}
}

// fetchOnce baut den Request mit den Discord-konformen Headern und führt ihn aus.
func (m *MediaDownloader) fetchOnce(
	ctx context.Context,
	originalURL, userAgent string,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, originalURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "image/*,*/*;q=0.8")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get failed: %w", err)
	}

	return resp, nil
}

// userAgentForURL liefert den von Discord vorgeschriebenen User-Agent für
// Discord-CDN-Hosts und einen generischen App-UA für alle anderen Quellen.
//
// Discord-Vorgabe: "DiscordBot ($url, $versionNumber)"
// https://docs.discord.com/developers/reference#user-agent
func (m *MediaDownloader) userAgentForURL(rawURL string) string {
	generic := "Billedapparat (" + appRepoURL + ", v" + m.version + ")"

	u, err := url.Parse(rawURL)
	if err != nil {
		return generic
	}

	switch {
	case strings.HasSuffix(u.Host, "discordapp.com"),
		strings.HasSuffix(u.Host, "discordapp.net"),
		strings.HasSuffix(u.Host, "discord.com"):
		return "DiscordBot (" + appRepoURL + ", v" + m.version + ")"
	default:
		return generic
	}
}

// retryDelay liest Retry-After bzw. X-RateLimit-Reset-After, sonst den Fallback.
func retryDelay(resp *http.Response, fallback time.Duration) time.Duration {
	for _, header := range []string{"Retry-After", "X-RateLimit-Reset-After"} {
		if v := resp.Header.Get(header); v != "" {
			if secs, err := strconv.ParseFloat(v, 64); err == nil {
				return clampWait(time.Duration(secs * float64(time.Second)))
			}
		}
	}

	return fallback
}

func backoffFor(attempt int) time.Duration {
	return clampWait(time.Duration(1<<uint(attempt-1)) * time.Second)
}

func clampWait(d time.Duration) time.Duration {
	if d > maxRetryWait {
		return maxRetryWait
	}

	return d
}

// sleepCtx wartet d, bricht aber bei Context-Cancellation sofort ab.
func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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
				Timeout:   dialTimeout,
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
