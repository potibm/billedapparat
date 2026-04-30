package cmd

import (
	"fmt"
	"log/slog"

	"github.com/potibm/billedapparat/internal/app/collectors"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/mastodon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCollectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect [source]",
		Short: "Start a collector to ingest slides from external sources (e.g. Mastodon, Bluesky)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]

			// 1. Get config for the specified collector source (z.B. collectors.mastodon)
			subViper := viper.Sub("collectors." + source)
			if subViper == nil {
				return fmt.Errorf("configuration for collector %s was not found", source)
			}

			if !subViper.GetBool("enabled") {
				return fmt.Errorf("collector %s is not enabled", source)
			}

			// 2. Get the Hub-URL and API-Key from the main config
			hubURL := viper.GetString("app.collector_url")
			if hubURL == "" {
				return fmt.Errorf("app.collector_url must be set in the configuration")
			}

			apiKey := subViper.GetString("api_key")
			if apiKey == "" {
				return fmt.Errorf("api_key missing for collector %s", source)
			}

			client := hubclient.New(hubURL, apiKey, slog.Default())

			var c collectors.Collector

			switch source {
			case "mastodon":
				var cfg mastodon.Config
				if err := subViper.Unmarshal(&cfg); err != nil {
					return fmt.Errorf("error parsing config for Mastodon Collector: %w", err)
				}

				c = mastodon.NewCollector(cfg, client)

			// case "bluesky": ...

			default:
				return fmt.Errorf("unknown collector source: %s", source)
			}

			slog.Info("Starting Collector", "source", source)

			ctx := cmd.Context()
			if err := c.Run(ctx); err != nil {
				return fmt.Errorf("collector error: %w", err)
			}

			slog.Info("Collector terminated", "source", source)

			return nil
		},
	}

	return cmd
}
