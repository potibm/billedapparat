package mastodon

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
	assert.Equal(t, "mastodon.social", cfg.Host)
	assert.Equal(t, "", cfg.AccessToken)
	assert.Equal(t, "demoscene", cfg.Tag)
}
