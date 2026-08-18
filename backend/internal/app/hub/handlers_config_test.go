package hub

import (
	"encoding/json"
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
		Beamer: config.BeamerConfig{
			AllowedAnimations: []string{"fade"},
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

	assert.Equal(t, []string{"fade"}, public.Beamer.AllowedAnimations)

	payload, err := json.Marshal(public)
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), "admin_api_key")

	// Regression guard: the frontend Zod schema reads snake_case keys.
	// Assert the marshaled JSON uses the exact key the schema expects, so
	// a missing `json:` tag on a config struct is caught here instead of
	// at runtime in the browser.
	assert.Contains(t, string(payload), `"allowed_animations":["fade"]`)
	assert.NotContains(t, string(payload), `"AllowedAnimations"`)
}
