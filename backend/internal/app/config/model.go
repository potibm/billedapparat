//nolint:lll // struct tags can get long and it's more readable to keep them in one line
package config

import "github.com/potibm/billedapparat/internal/app/domain"

type OrderType string

const (
	OrderRandom OrderType = "random"
	OrderAsc    OrderType = "asc"
	OrderDesc   OrderType = "desc"
)

type PlaylistStep struct {
	Type     domain.SlideType `json:"type"     yaml:"type"`
	Order    OrderType        `json:"order"    yaml:"order"`
	Count    int              `json:"count"    yaml:"count"`
	Duration int              `json:"duration" yaml:"duration"`
}

type PlaylistConfig struct {
	ID    int            `json:"id"    yaml:"id"`
	Name  string         `json:"name"  yaml:"name"`
	Steps []PlaylistStep `json:"steps" yaml:"steps"`
}

type SentryConfig struct {
	DSN                     string  `json:"dsn"                        mapstructure:"dsn"                        validate:"omitempty,url"`
	TraceSampleRate         float64 `json:"trace_sample_rate"          mapstructure:"trace_sample_rate"          validate:"omitempty,gte=0,lte=1"`
	ReplaySessionSampleRate float64 `json:"replay_session_sample_rate" mapstructure:"replay_session_sample_rate" validate:"omitempty,gte=0,lte=1"`
	ReplayErrorSampleRate   float64 `json:"replay_error_sample_rate"   mapstructure:"replay_error_sample_rate"   validate:"omitempty,gte=0,lte=1"`
	Environment             string  `json:"environment"                mapstructure:"environment"                validate:"required"`
	Version                 string  `json:"version"                    mapstructure:"version"                    validate:"required"`
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

	OtelEndpoint string `mapstructure:"otel_endpoint" validate:"omitempty"`
	Port         int    `mapstructure:"port"          validate:"required,gt=0,lte=65535"`
}

type FormatConfig struct {
	Date DateFormatConfig `json:"date" mapstructure:"date"`
}

type DateFormatOptionsConfig map[string]any

type DateFormatConfig struct {
	Locale  string                  `json:"locale"  mapstructure:"locale"  validate:"required"`
	Options DateFormatOptionsConfig `json:"options" mapstructure:"options"`
}

type CorsAllowOriginsConfig []string

type APIConfig struct {
	AdminAPIKey string `mapstructure:"admin_api_key" validate:"required"`
}

type CollectorConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key" validate:"required_if=Enabled true"`
}

type SyncConfig struct {
	RedisURL    RedisURL `mapstructure:"redis_url"    validate:"url"`
	NewsStream  string   `mapstructure:"news_stream"`
	EventStream string   `mapstructure:"event_stream"`
}

type Config struct {
	App        AppConfig                  `mapstructure:"app"`
	Format     FormatConfig               `mapstructure:"format"`
	Sentry     SentryConfig               `mapstructure:"sentry"`
	API        APIConfig                  `mapstructure:"api"`
	Collectors map[string]CollectorConfig `mapstructure:"collectors" validate:"dive"`
	Playlists  []PlaylistConfig           `mapstructure:"playlists"`
	Sync       SyncConfig                 `mapstructure:"sync"`
}

func (p *PlaylistStep) SetDefaults() {
	if p.Order == "" {
		p.Order = OrderRandom
	}

	if p.Count <= 0 {
		p.Count = 1
	}

	if p.Duration <= 0 {
		p.Duration = 10
	}
}
