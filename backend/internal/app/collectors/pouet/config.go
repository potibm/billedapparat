package pouet

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

const pouetDefaultPollIntervalOneHour = 60

type Keywords []string

type Config struct {
	config.CollectorConfig `mapstructure:",squash"`

	PollInterval int      `mapstructure:"poll_interval"`
	Keywords     Keywords `mapstructure:"keywords"`
}

func (k Keywords) Lower() Keywords {
	lowerKeywords := make(Keywords, len(k))
	for i, keyword := range k {
		lowerKeywords[i] = strings.ToLower(keyword)
	}

	return lowerKeywords
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
