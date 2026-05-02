package hub

import (
	"testing"

	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestMapToPublicConfig(t *testing.T) {
	internalCfg := &config.Config{
		App: config.AppConfig{
			Version:            "1.2.3",
			Environment:        "production",
			EnvironmentMessage: "Hello World",
		},
		Sentry: config.SentryConfig{
			DSN:         "https://secret@sentry.io/123",
			Environment: "prod",
			Version:     "v1",
		},
		API: config.APIConfig{
			AdminAPIKey: "SECRET_KEY_NEVER_REVEAL",
		},
		Playlists: []config.PlaylistConfig{
			{ID: 1, Name: "Test Playlist"},
		},
	}

	public := mapToPublicConfig(internalCfg)

	// Verification
	assert.Equal(t, "1.2.3", public.Version)
	assert.Equal(t, "production", public.Environment)
	assert.Equal(t, "Hello World", public.EnvironmentMessage)
	assert.Equal(t, "https://secret@sentry.io/123", public.Sentry.DSN)

	assert.Len(t, public.Playlists, 1)
	assert.Equal(t, "Test Playlist", public.Playlists[0].Name)
}
