package mastodon

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

type Config struct {
	config.CollectorConfig `mapstructure:",squash"`

	Host        string `mapstructure:"host"         validate:"required"`
	AccessToken string `mapstructure:"access_token" validate:"required"`
	Tag         string `mapstructure:"tag"          validate:"required"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeSlide,
			APIKey:  generatedAPIKey,
		},
		Host:        "mastodon.social",
		AccessToken: "",
		Tag:         "demoscene",
	}

	var result map[string]any

	_ = mapstructure.Decode(configStruct, &result)

	return result
}
