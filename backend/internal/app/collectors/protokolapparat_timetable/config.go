package protokolapparat_timetable

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
)

const (
	defaultStreamName    = "party:timetable:events"
	defaultConsumerGroup = "billedapparat_timetable_collector"
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
			Type:    config.CollectorDataTypeTimetable,
			APIKey:  generatedAPIKey,
		},
		RedisURL:      "redis://localhost:6379",
		StreamName:    defaultStreamName,
		ConsumerGroup: defaultConsumerGroup,
		ConsumerName:  defaultConsumerName,
	}

	var result map[string]any

	_ = mapstructure.Decode(configStruct, &result)

	return result
}
