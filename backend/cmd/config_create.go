package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

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

			viper.Set("api.admin_api_key", generateSecureToken(defaultAPIKeyLength))

			viper.Set("format.date.locale", "de-DE")
			viper.Set("format.date.options", config.DateFormatOptionsConfig{
				"timeZone": "Europe/Berlin",
				"weekday":  "long",
				"year":     "numeric",
				"month":    "long",
				"day":      "numeric",
			})

			collectors := map[string]any{
				"mastodon": mastodon.DefaultConfig(generateSecureToken(32)),
			}
			viper.Set("collectors", collectors)

			filename := "config.yaml"

			err := viper.SafeWriteConfigAs(filename)
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

func generateSecureToken(byteLength int) string {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Error while generating the token: %v", err)
	}

	return hex.EncodeToString(bytes)
}
