package protokolapparat_news

import "github.com/potibm/billedapparat/internal/app/config"

type Config struct {
	Enabled    bool                     `mapstructure:"enabled"`
	Type       config.CollectorDataType `mapstructure:"type"`
	APIKey     string                   `mapstructure:"api_key"`
	RedisURL   config.RedisURL          `mapstructure:"redis_url"`
	StreamName string                   `mapstructure:"stream_name"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	return map[string]any{
		"enabled":     false,
		"type":        config.CollectorDataTypeNews,
		"api_key":     generatedAPIKey,
		"redis_url":   "redis://localhost:6379",
		"stream_name": "party:news:events",
	}
}
