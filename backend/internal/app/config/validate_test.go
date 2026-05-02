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

func TestPlaylistConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  PlaylistConfig
		wantErr bool
	}{
		{
			name: "valid playlist",
			config: PlaylistConfig{
				ID:   1,
				Name: "Valid Playlist",
				Steps: []PlaylistStep{
					{Type: "sponsor"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: PlaylistConfig{
				ID:    2,
				Name:  "",
				Steps: []PlaylistStep{{Type: "sponsor"}},
			},
			wantErr: true,
		},
		{
			name: "no steps",
			config: PlaylistConfig{
				ID:    3,
				Name:  "No Steps",
				Steps: []PlaylistStep{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDateFormatConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DateFormatConfig
		wantErr bool
	}{
		{
			name: "valid locale",
			config: DateFormatConfig{
				Locale: "en-US",
			},
			wantErr: false,
		},
		{
			name: "invalid locale",
			config: DateFormatConfig{
				Locale: "invalid-locale",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatConfig_Validate(t *testing.T) {
	cfg := FormatConfig{
		Date: DateFormatConfig{
			Locale: "en-US",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestAppConfig_Validate(t *testing.T) {
	cfg := AppConfig{
		GinMode:      "debug",
		Environment:  "development",
		LogLevel:     "info",
		LogFormat:    "text",
		DbFilename:   "test.db",
		FrontendURL:  "http://localhost:3000",
		CollectorURL: "http://localhost:8080",
	}

	err := cfg.Validate()
	assert.NoError(t, err)

	cfg.DbFilename = "../invalid-filename"
	err = cfg.Validate()
	assert.Error(t, err)
}

func TestAPIConfig_Validate(t *testing.T) {
	cfg := APIConfig{
		AdminAPIKey: "test-key-12345",
	}

	err := cfg.Validate("production")
	assert.NoError(t, err)

	cfg.AdminAPIKey = DefaultAPIAdminKey
	err = cfg.Validate("production")
	assert.Error(t, err)

	err = cfg.Validate("development")
	assert.NoError(t, err)
}
