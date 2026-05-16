package protokolapparat_news

import (
	"encoding/json"

	"github.com/potibm/billedapparat/internal/app/config"
)

const (
	defaultStreamName    = "party:news:events"
	defaultConsumerGroup = "billedapparat_news_collector"
	defaultConsumerName  = "instance_1"
)

type Config struct {
	config.CollectorConfig

	RedisURL      config.RedisURL `mapstructure:"redis_url"`
	StreamName    string          `mapstructure:"stream_name"`
	ConsumerGroup string          `mapstructure:"consumer_group" validate:"required"`
	ConsumerName  string          `mapstructure:"consumer_name"  validate:"required"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	configStruct := Config{
		CollectorConfig: config.CollectorConfig{
			Enabled: false,
			Type:    config.CollectorDataTypeNews,
			APIKey:  generatedAPIKey,
		},
		RedisURL:      "redis://localhost:6379",
		StreamName:    defaultStreamName,
		ConsumerGroup: defaultConsumerGroup,
		ConsumerName:  defaultConsumerName,
	}

	var result map[string]any

	data, _ := json.Marshal(configStruct)
	_ = json.Unmarshal(data, &result)

	return result
}
