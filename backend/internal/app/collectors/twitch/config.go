package twitch

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

type Config struct {
	config.CollectorConfig `mapstructure:",squash"`

	Channel      string `mapstructure:"channel"       validate:"required"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeSlide,
			APIKey:  generatedAPIKey,
		},
		Channel:      "endless_demoshow",
		ClientID:     "",
		ClientSecret: "",
	}

	var result map[string]any

	_ = mapstructure.Decode(configStruct, &result)

	return result
}
