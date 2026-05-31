package bluesky

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

type Hashtags []string

type Config struct {
	config.CollectorConfig

	Hashtags Hashtags `mapstructure:"hashtags"`
}

func (h Hashtags) Lower() Hashtags {
	lowerHashtags := make(Hashtags, len(h))
	for i, hashtag := range h {
		// @todo add a # if it's missing?
		lowerHashtags[i] = strings.ToLower(hashtag)
	}

	return lowerHashtags
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeSlide,
			APIKey:  generatedAPIKey,
		},
		Hashtags: []string{"#demoscene"},
	}

	var result map[string]any

	_ = mapstructure.Decode(configStruct, &result)

	return result
}
