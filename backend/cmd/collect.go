package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/mastodon"
	"github.com/potibm/billedapparat/internal/app/collectors/pouet"
	"github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_news"
	"github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_timetable"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/initializer"
	"github.com/redis/go-redis/v9"
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

	c, err := buildCollector(source, subViper, client)
	if err != nil {
		return err
	}

	slog.Info("Starting Collector", "source", source)

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

func buildCollector(source string, subViper *viper.Viper, client *hubclient.HubClient) (collectors.Collector, error) {
	switch source {
	case "mastodon":
		var cfg mastodon.Config
		if err := subViper.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("error parsing config for Mastodon Collector: %w", err)
		}

		return mastodon.NewCollector(cfg, client), nil

	case "pouet":
		var cfg pouet.Config
		if err := subViper.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("error parsing config for Pouet Collector: %w", err)
		}

		return pouet.NewCollector(cfg, client), nil

	case "protokolapparat-news":
		var cfg protokolapparat_news.Config
		if err := subViper.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("error parsing config for Protokolapparat News Collector: %w", err)
		}

		rdb, err := initializeRedisClient(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("error initializing Redis client: %w", err)
		}

		return protokolapparat_news.NewCollector(cfg, client, rdb), nil

	case "protokolapparat-timetable":
		var cfg protokolapparat_timetable.Config
		if err := subViper.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("error parsing config for Protokolapparat Timetable Collector: %w", err)
		}

		rdb, err := initializeRedisClient(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("error initializing Redis client: %w", err)
		}

		return protokolapparat_timetable.NewCollector(cfg, client, rdb), nil

	default:
		return nil, fmt.Errorf("unknown collector source: %s", source)
	}
}

func initializeRedisClient(redisURL config.RedisURL) (*redis.Client, error) {
	options, err := redis.ParseURL(string(redisURL))
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	rdb := redis.NewClient(options)

	// Test connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("could not connect to Redis: %w", err)
	}

	return rdb, nil
}
