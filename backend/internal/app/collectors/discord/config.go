package discord

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

type Config struct {
	config.CollectorConfig

	BotToken  string `mapstructure:"bot_token"`
	ChannelID string `mapstructure:"channel_id"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeSlide,
			APIKey:  generatedAPIKey,
		},
		BotToken:  "",
		ChannelID: "0",
	}

	var result map[string]any

	_ = mapstructure.Decode(configStruct, &result)

	return result
}
