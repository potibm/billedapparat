package mastodon

import (
	"encoding/json"

	"github.com/potibm/billedapparat/internal/app/config"
)

type Config struct {
	config.CollectorConfig

	Host        string `mapstructure:"host"`
	AccessToken string `mapstructure:"access_token"`
	Tag         string `mapstructure:"tag"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeNews,
			APIKey:  generatedAPIKey,
		},
		Host:        "mastodon.social",
		AccessToken: "",
		Tag:         "demoscene",
	}

	var result map[string]any

	data, _ := json.Marshal(configStruct)
	_ = json.Unmarshal(data, &result)

	return result
}
