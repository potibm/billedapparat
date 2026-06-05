package protokolapparat_news

import (
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	apiKey := "test-api-key"
	defaults := DefaultConfig(apiKey)

	var cfg Config

	err := mapstructure.Decode(defaults, &cfg)

	assert.NoError(t, err)
	assert.False(t, cfg.Enabled)
	assert.Equal(t, apiKey, cfg.APIKey)
	assert.Equal(t, config.CollectorDataTypeNews, cfg.Type)
	assert.Equal(t, config.RedisURL("redis://localhost:6379"), cfg.RedisURL)
	assert.Equal(t, defaultStreamName, cfg.StreamName)
	assert.Equal(t, defaultConsumerGroup, cfg.ConsumerGroup)
	assert.Equal(t, defaultConsumerName, cfg.ConsumerName)
}
