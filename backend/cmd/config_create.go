package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/potibm/billedapparat/internal/app/collectors/mastodon"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConfigCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new configuration file with default values",
		RunE: func(cmd *cobra.Command, args []string) error {
			const defaultAPIKeyLength = 32

			adminAPIKey, err := generateSecureToken(defaultAPIKeyLength)
			if err != nil {
				return fmt.Errorf("failed to generate admin api key: %w", err)
			}

			viper.Set("api.admin_api_key", adminAPIKey)

			viper.Set("format.date.locale", "de-DE")
			viper.Set("format.date.options", config.DateFormatOptionsConfig{
				"timeZone": "Europe/Berlin",
				"weekday":  "long",
				"year":     "numeric",
				"month":    "long",
				"day":      "numeric",
			})

			mastodonCollectorAPIKey, err := generateSecureToken(defaultAPIKeyLength)
			if err != nil {
				return fmt.Errorf("failed to generate collector api key: %w", err)
			}

			collectors := map[string]any{
				"mastodon": mastodon.DefaultConfig(mastodonCollectorAPIKey),
			}
			viper.Set("collectors", collectors)

			filename := "config.yaml"

			err = viper.SafeWriteConfigAs(filename)
			if err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return fmt.Errorf(
						"file %s already exists or was not able to be created: %w",
						filename,
						err,
					)
				}

				return fmt.Errorf("error writing the confiig: %w", err)
			}

			fmt.Printf("✅ Successfully created: %s\n", filename)

			return nil
		},
	}
}

func generateSecureToken(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error while generating the token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
