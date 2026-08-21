package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	OtelServiceName        = "billedapparat"
	OtelBackendServiceName = OtelServiceName + "-backend"

	DefaultPort = 8080

	DefaultTraceSampleRate         = 0.1
	DefaultReplaySessionSampleRate = 0.1
	DefaultReplayErrorSampleRate   = 0.1
	DefaultJwtSecret               = "very-insecure"
	DefaultAPIAdminKey             = "also-very-insecure"

	DataDirname      = "./data"
	DatabaseDirname  = DataDirname + "/db"
	MediaDirname     = DataDirname + "/media"
	StylesDirname    = DataDirname + "/style"
	ImportDirname    = DataDirname + "/import"
	SeedCacheDirname = DataDirname + "/seed_cache"

	MediaURL  = "/media/"
	StylesURL = "/style/"

	DefaultDBFilename = "billedapparat"

	DataDirPerm = 0o755
)

var DefaultDateOptions = DateFormatOptionsConfig{
	"weekday": "long",
	"hour":    "2-digit",
	"minute":  "2-digit",
}

func InitViper() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("app.port", DefaultPort)
	viper.SetDefault("app.otel_endpoint", "")
	viper.SetDefault("app.gin_mode", "release")
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.env", "production")
	viper.SetDefault("app.db_filename", DefaultDBFilename)
	viper.SetDefault("app.frontend_url", "")
	viper.SetDefault("app.collector_url", "")
	viper.SetDefault("app.cors_allow_origins", []string{})
	viper.SetDefault("api.admin_api_key", DefaultAPIAdminKey)

	viper.SetDefault("format.date.locale", "da-DK")
	viper.SetDefault("format.date.options", DefaultDateOptions)

	viper.SetDefault("sentry.dsn", "")
	viper.SetDefault("sentry.trace_sample_rate", DefaultTraceSampleRate)
	viper.SetDefault("sentry.replay_session_sample_rate", DefaultReplaySessionSampleRate)
	viper.SetDefault("sentry.replay_error_sample_rate", DefaultReplayErrorSampleRate)

	viper.SetDefault("admin_urls.timetable", "")
	viper.SetDefault("admin_urls.news", "")

	viper.SetDefault("timetable.max_entries_per_slide", DefaultTimetableMaxEntriesPerSlide)
	viper.SetDefault("timetable.timezone", "UTC")

	viper.SetDefault("beamer.allowed_animations", []string{"fade", "slideRight", "zoomIn", "flip", "urgent"})

	viper.RegisterAlias("sentry.environment", "app.env")
	viper.RegisterAlias("sentry.version", "app.version")

	viper.SetDefault("playlists", []PlaylistConfig{
		{
			ID:   1,
			Name: "Default",
			Steps: []PlaylistStep{
				//nolint:mnd // reasonable defaults for playlist steps
				{Type: "sponsor", Order: OrderRandom, Count: 1, Duration: 10},
			},
		},
	})
}
