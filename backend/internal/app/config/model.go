package config

type SentryConfig struct {
	DSN                     string  `mapstructure:"dsn"                        validate:"omitempty,url"`
	TraceSampleRate         float64 `mapstructure:"trace_sample_rate"          validate:"omitempty,gte=0,lte=1"`
	ReplaySessionSampleRate float64 `mapstructure:"replay_session_sample_rate" validate:"omitempty,gte=0,lte=1"`
	ReplayErrorSampleRate   float64 `mapstructure:"replay_error_sample_rate"   validate:"omitempty,gte=0,lte=1"`
	Environment             string  `mapstructure:"environment"`
	Version                 string  `mapstructure:"version"`
}

type AppConfig struct {
	Version string `mapstructure:"version"`

	GinMode     string `mapstructure:"gin_mode" validate:"required,oneof=debug release test"`
	Environment string `mapstructure:"env"      validate:"required,oneof=development staging production test"`

	LogLevel  string `mapstructure:"log_level"  validate:"required,oneof=debug info warn error"`
	LogFormat string `mapstructure:"log_format" validate:"required,oneof=json text"`

	DbFilename         string                 `mapstructure:"db_filename"         validate:"required"`
	FrontendURL        string                 `mapstructure:"frontend_url"        validate:"required,http_url"`
	CollectorURL       string                 `mapstructure:"collector_url"       validate:"required,http_url"`
	CorsAllowOrigins   CorsAllowOriginsConfig `mapstructure:"cors_allow_origins"  validate:"dive,required"`
	EnvironmentMessage string                 `mapstructure:"environment_message"`
}

type FormatConfig struct {
	Date DateFormatConfig `mapstructure:"date"`
}

type DateFormatOptionsConfig map[string]any

type DateFormatConfig struct {
	Locale  string                  `mapstructure:"locale"  validate:"required"`
	Options DateFormatOptionsConfig `mapstructure:"options"`
}

type CorsAllowOriginsConfig []string

type APIConfig struct {
	AdminAPIKey string `mapstructure:"admin_api_key" validate:"required"`
}

type CollectorConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key" validate:"required_if=Enabled true"`
}

type Config struct {
	App        AppConfig                  `mapstructure:"app"`
	Format     FormatConfig               `mapstructure:"format"`
	Sentry     SentryConfig               `mapstructure:"sentry"`
	API        APIConfig                  `mapstructure:"api"`
	Collectors map[string]CollectorConfig `mapstructure:"collectors" validate:"dive"`
}
