package pouet

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

const pouetDefaultPollIntervalOneHour = 60

type Config struct {
	config.CollectorConfig

	PollInterval int      `mapstructure:"poll_interval"`
	Keywords     []string `mapstructure:"keywords"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeSlide,
			APIKey:  generatedAPIKey,
		},
		PollInterval: pouetDefaultPollIntervalOneHour,
		Keywords:     []string{"demoscene"},
	}

	var result map[string]any

	_ = mapstructure.Decode(configStruct, &result)

	return result
}
