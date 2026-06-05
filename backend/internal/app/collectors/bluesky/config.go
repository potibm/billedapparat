package bluesky

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

type Hashtags []string

type Config struct {
	config.CollectorConfig `mapstructure:",squash"`

	Hashtags Hashtags `mapstructure:"hashtags" validate:"required,dive,required"`
}

func (h Hashtags) Normalize() Hashtags {
	normalized := make(Hashtags, 0, len(h))
	for _, tag := range h {
		clean := strings.ToLower(strings.TrimPrefix(tag, "#"))
		if clean != "" {
			normalized = append(normalized, clean)
		}
	}

	return normalized
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeSlide,
			APIKey:  generatedAPIKey,
		},
		Hashtags: Hashtags{"#demoscene"},
	}

	var result map[string]any

	_ = mapstructure.Decode(configStruct, &result)

	return result
}
