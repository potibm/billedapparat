package hub

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMediaDownloader(t *testing.T, client *http.Client) *MediaDownloader {
	t.Helper()

	dl := newMediaDownloader(nil, nil, slog.New(slog.DiscardHandler), "test", client)
	dl.retryBackoff = func(int) time.Duration { return time.Millisecond }

	return dl
}

func TestUserAgentForURL(t *testing.T) {
	t.Parallel()

	dl := newTestMediaDownloader(t, &http.Client{})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "discord cdn avatar",
			url:  "https://cdn.discordapp.com/avatars/123/hash.png?size=128",
			want: "DiscordBot (https://github.com/potibm/billedapparat, vtest)",
		},
		{
			name: "discord cdn attachment",
			url:  "https://cdn.discordapp.com/attachments/1/2/image.png",
			want: "DiscordBot (https://github.com/potibm/billedapparat, vtest)",
		},
		{
			name: "discord media net",
			url:  "https://media.discordapp.net/stickers/1.gif",
			want: "DiscordBot (https://github.com/potibm/billedapparat, vtest)",
		},
		{
			name: "other host",
			url:  "https://files.mastodon.social/media/1.png",
			want: "Billedapparat (https://github.com/potibm/billedapparat, vtest)",
		},
		{
			name: "invalid url falls back to generic",
			url:  "://broken",
			want: "Billedapparat (https://github.com/potibm/billedapparat, vtest)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, dl.userAgentForURL(tt.url))
		})
	}
}

func TestIsRetryableStatus(t *testing.T) {
	t.Parallel()

	retryable := []int{
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}
	for _, code := range retryable {
		assert.True(t, isRetryableStatus(code), "status %d should be retryable", code)
	}

	nonRetryable := []int{
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusBadRequest,
		http.StatusTooManyRequests, // wird separat behandelt
	}
	for _, code := range nonRetryable {
		assert.False(t, isRetryableStatus(code), "status %d must not be retryable", code)
	}
}

func TestFetchWithRetry_SendsDiscordUserAgent(t *testing.T) {
	t.Parallel()

	var gotUA string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Dial umbiegen, damit der Original-Host (cdn.discordapp.com) erhalten bleibt und
	// der Discord-User-Agent greift, der Request aber lokal ankommt.
	client := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, srv.Listener.Addr().String())
		},
	}}

	dl := newTestMediaDownloader(t, client)

	// http statt https: der Testserver terminiert kein TLS; der Host bleibt ausschlaggebend.
	resp, err := dl.fetchWithRetry(context.Background(), "http://cdn.discordapp.com/avatars/1/x.png")
	require.NoError(t, err)

	resp.Body.Close()

	assert.Equal(t, "DiscordBot (https://github.com/potibm/billedapparat, vtest)", gotUA)
}

func TestFetchWithRetry_NoRetryOnPermanentClientErrors(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int64

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
			}))
			defer srv.Close()

			dl := newTestMediaDownloader(t, srv.Client())

			resp, err := dl.fetchWithRetry(context.Background(), srv.URL)
			if resp != nil {
				resp.Body.Close()
			}

			require.Error(t, err)
			assert.EqualValues(t, 1, calls.Load(), "status %d must not be retried", status)
		})
	}
}

func TestFetchWithRetry_RetriesTransientServerErrors(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dl := newTestMediaDownloader(t, srv.Client())

	resp, err := dl.fetchWithRetry(context.Background(), srv.URL)
	require.NoError(t, err)

	resp.Body.Close()

	assert.EqualValues(t, 3, calls.Load())
}

func TestFetchWithRetry_GivesUpAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	dl := newTestMediaDownloader(t, srv.Client())

	resp, err := dl.fetchWithRetry(context.Background(), srv.URL)
	if resp != nil {
		resp.Body.Close()
	}

	require.Error(t, err)
	assert.EqualValues(t, maxFetchAttempts, calls.Load())
}

func TestFetchWithRetry_RetriesOn429WithRetryAfter(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dl := newTestMediaDownloader(t, srv.Client())

	resp, err := dl.fetchWithRetry(context.Background(), srv.URL)
	require.NoError(t, err)

	resp.Body.Close()

	assert.EqualValues(t, 2, calls.Load())
}

func TestFetchWithRetry_AbortsOnContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dl := newTestMediaDownloader(t, &http.Client{})

	resp, err := dl.fetchWithRetry(ctx, "https://cdn.discordapp.com/avatars/1/x.png")
	if resp != nil {
		resp.Body.Close()
	}

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCachedAvatarURL_CachesResult(t *testing.T) {
	t.Parallel()

	dl := newTestMediaDownloader(t, &http.Client{})

	var downloads atomic.Int64

	load := func(context.Context, string) (string, error) {
		downloads.Add(1)

		return "/media/avatar.webp", nil
	}

	first, err := dl.cachedAvatarURL(context.Background(), "https://cdn.discordapp.com/avatars/1/a.png", load)
	require.NoError(t, err)

	second, err := dl.cachedAvatarURL(context.Background(), "https://cdn.discordapp.com/avatars/1/a.png", load)
	require.NoError(t, err)

	assert.Equal(t, "/media/avatar.webp", first)
	assert.Equal(t, first, second)
	assert.EqualValues(t, 1, downloads.Load(), "avatar must be downloaded once")
}

func TestCachedAvatarURL_DeduplicatesConcurrentCalls(t *testing.T) {
	t.Parallel()

	dl := newTestMediaDownloader(t, &http.Client{})

	var downloads atomic.Int64

	load := func(context.Context, string) (string, error) {
		downloads.Add(1)
		time.Sleep(10 * time.Millisecond)

		return "/media/avatar.webp", nil
	}

	const goroutines = 20

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			_, err := dl.cachedAvatarURL(context.Background(), "https://cdn.discordapp.com/avatars/1/a.png", load)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	assert.EqualValues(t, 1, downloads.Load(), "concurrent calls must share one download")
}

func TestCachedAvatarURL_DoesNotCacheErrors(t *testing.T) {
	t.Parallel()

	dl := newTestMediaDownloader(t, &http.Client{})

	var downloads atomic.Int64

	load := func(context.Context, string) (string, error) {
		downloads.Add(1)

		return "", errors.New("boom")
	}

	for range 2 {
		_, err := dl.cachedAvatarURL(context.Background(), "https://cdn.discordapp.com/avatars/1/a.png", load)
		require.Error(t, err)
	}

	assert.EqualValues(t, 2, downloads.Load(), "failures must not populate the cache")
}

func TestCachedAvatarURL_RejectsEmptyURL(t *testing.T) {
	t.Parallel()

	dl := newTestMediaDownloader(t, &http.Client{})

	_, err := dl.cachedAvatarURL(context.Background(), "", func(context.Context, string) (string, error) {
		t.Fatal("download must not be called for an empty url")

		return "", nil
	})

	require.Error(t, err)
}

func TestRetryDelay(t *testing.T) {
	t.Parallel()

	fallback := 3 * time.Second

	resp := &http.Response{Header: http.Header{}}
	assert.Equal(t, fallback, retryDelay(resp, fallback))

	resp.Header.Set("Retry-After", "2")
	assert.Equal(t, 2*time.Second, retryDelay(resp, fallback))

	resp.Header.Del("Retry-After")
	resp.Header.Set("X-RateLimit-Reset-After", "1.5")
	assert.Equal(t, 1500*time.Millisecond, retryDelay(resp, fallback))

	resp.Header.Set("Retry-After", "3600")
	assert.Equal(t, maxRetryWait, retryDelay(resp, fallback), "retry-after must be clamped")
}

func TestBackoffFor(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 1*time.Second, backoffFor(1))
	assert.Equal(t, 2*time.Second, backoffFor(2))
	assert.Equal(t, 4*time.Second, backoffFor(3))
	assert.Equal(t, maxRetryWait, backoffFor(20))
}
