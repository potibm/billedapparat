package mastodon

type Config struct {
	Enabled     bool   `mapstructure:"enabled"`
	APIKey      string `mapstructure:"api_key"`
	Host        string `mapstructure:"host"`
	AccessToken string `mapstructure:"access_token"`
	Tag         string `mapstructure:"tag"`
}

func DefaultConfig(generatedAPIKey string) map[string]any {
	return map[string]any{
		"enabled":      false,
		"api_key":      generatedAPIKey,
		"host":         "mastodon.social",
		"access_token": "",
		"tag":          "demoscene",
	}
}
