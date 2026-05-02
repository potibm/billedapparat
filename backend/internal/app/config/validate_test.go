package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_PlaylistDefaultsAndValidation(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			GinMode:      "debug",
			Environment:  "development",
			LogLevel:     "info",
			LogFormat:    "text",
			DbFilename:   "test.db",
			FrontendURL:  "http://localhost:3000",
			CollectorURL: "http://localhost:8080",
		},
		Format: FormatConfig{
			Date: DateFormatConfig{
				Locale: "da-DK",
			},
		},
		API: APIConfig{
			AdminAPIKey: "test-key-12345",
		},
		Sentry: SentryConfig{
			DSN:                     "https://test@sentry.io/123",
			TraceSampleRate:         0.1,
			ReplaySessionSampleRate: 0.1,
			ReplayErrorSampleRate:   0.1,
			Environment:             "development",
			Version:                 "1.2.3",
		},
		Playlists: []PlaylistConfig{
			{
				ID:   1,
				Name: "Beamer Loop",
				Steps: []PlaylistStep{
					{Type: "sponsor"},
				},
			},
		},
	}

	// 1. trigger validation
	err := cfg.Validate()
	assert.NoError(t, err)

	// 2. Check on defaults
	step := cfg.Playlists[0].Steps[0]
	assert.Equal(t, OrderRandom, step.Order)
	assert.Equal(t, 1, step.Count)
	assert.Equal(t, 10, step.Duration)
}
