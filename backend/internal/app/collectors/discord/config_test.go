package discord

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
	assert.Equal(t, config.CollectorDataTypeSlide, cfg.Type)
	assert.Equal(t, "", cfg.BotToken)
	assert.Equal(t, "0", cfg.ChannelID)
	assert.False(t, cfg.DropAttachments)
}
