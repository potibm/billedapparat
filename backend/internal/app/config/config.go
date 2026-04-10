package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	OtelServiceName        = "billedapparat"
	OtelBackendServiceName = OtelServiceName + "-backend"

	DefaultTraceSampleRate         = 0.1
	DefaultReplaySessionSampleRate = 0.1
	DefaultReplayErrorSampleRate   = 0.1
	DefaultJwtSecret               = "very-insecure"

	DefaultDBFilename = "billedapparat"
)

var (
	DefaultDateOptions = DateFormatOptionsConfig{
		"weekday": "long",
		"hour":    "2-digit",
		"minute":  "2-digit",
	}
)

func InitViper() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("app.gin_mode", "release")
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.env", "production")
	viper.SetDefault("app.db_filename", DefaultDBFilename)
	viper.SetDefault("app.frontend_url", "")
	viper.SetDefault("app.cors_allow_origins", []string{})

	viper.SetDefault("format.date.locale", "da-DK")
	viper.SetDefault("format.date.options", DefaultDateOptions)

	viper.SetDefault("sentry.dsn", "")
	viper.SetDefault("sentry.trace_sample_rate", DefaultTraceSampleRate)
	viper.SetDefault("sentry.replay_session_sample_rate", DefaultReplaySessionSampleRate)
	viper.SetDefault("sentry.replay_error_sample_rate", DefaultReplayErrorSampleRate)

	viper.RegisterAlias("sentry.environment", "app.env")
	viper.RegisterAlias("sentry.version", "app.version")
}
