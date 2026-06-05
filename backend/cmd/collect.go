package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/factory"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/initializer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	waitForServerMaxRetries   = 10
	waitForServerInitialDelay = 1 * time.Second
)

func NewCollectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect [source]",
		Short: "Start a collector to ingest slides from external sources (e.g. Mastodon, Bluesky)",
		Args:  cobra.ExactArgs(1),
		RunE:  runCollectorCommand,
	}

	return cmd
}

func runCollectorCommand(cmd *cobra.Command, args []string) error {
	// 1. Context
	source := args[0]
	ctx := cmd.Context()

	// 2. Initialize Telemetry
	shutdownFn, err := initializer.InitTelemetry(ctx, Cfg.App.OtelEndpoint, Cfg.App.Version)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	if shutdownFn != nil {
		defer shutdownFn()
	}

	meter := otel.Meter(config.OtelMeterName)

	runningCollectorCounter, err := meter.Int64UpDownCounter(
		"billedapparat_collector_running",
		metric.WithDescription("Running collector processes"),
	)
	if err != nil {
		slog.Warn("Failed to initialize running collector counter", "err", err)
	}

	// 3. Set up Collector based on source
	subViper, hubURL, apiKey, err := loadCollectorConfig(source)
	if err != nil {
		return err
	}

	client := hubclient.New(hubURL, apiKey, slog.Default())

	err = client.WaitForServer(ctx, waitForServerMaxRetries, waitForServerInitialDelay)
	if err != nil {
		return fmt.Errorf("failed to wait for hub server: %w", err)
	}

	c, err := factory.Build(source, subViper, client)
	if err != nil {
		return err
	}

	slog.Info("Starting Collector", "source", source)

	setTerminalTitle(fmt.Sprintf("📡 Billedapparat Collector - %s", source))

	defer func() {
		if err := c.Close(); err != nil {
			slog.Error("Failed to close collector", "err", err)
		}
	}()

	counterAttrs := metric.WithAttributes(
		attribute.String("source", source),
	)
	if runningCollectorCounter != nil {
		runningCollectorCounter.Add(ctx, 1, counterAttrs)
		defer runningCollectorCounter.Add(ctx, -1, counterAttrs)
	}

	if err := c.Run(ctx); err != nil {
		return fmt.Errorf("collector error: %w", err)
	}

	slog.Info("Collector terminated", "source", source)

	clearTerminal()

	return nil
}

func loadCollectorConfig(source string) (subViper *viper.Viper, hubURL, apiKey string, err error) {
	subViper = viper.Sub("collectors." + source)
	if subViper == nil {
		return nil, "", "", fmt.Errorf("configuration for collector %s was not found", source)
	}

	if !subViper.GetBool("enabled") {
		return nil, "", "", fmt.Errorf("collector %s is not enabled", source)
	}

	hubURL = viper.GetString("app.collector_url")
	if hubURL == "" {
		return nil, "", "", fmt.Errorf("app.collector_url must be set in the configuration")
	}

	apiKey = subViper.GetString("api_key")
	if apiKey == "" {
		return nil, "", "", fmt.Errorf("api_key missing for collector %s", source)
	}

	return subViper, hubURL, apiKey, nil
}
