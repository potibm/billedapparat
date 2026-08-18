package lorem

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

const loremDefaultPollIntervalSeconds = 15

type Keywords []string

type Config struct {
	config.CollectorConfig `mapstructure:",squash"`

	PollInterval int `mapstructure:"poll_interval"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeSlide,
			APIKey:  generatedAPIKey,
		},
		PollInterval: loremDefaultPollIntervalSeconds,
	}

	var result map[string]any

	_ = mapstructure.Decode(configStruct, &result)

	return result
}
