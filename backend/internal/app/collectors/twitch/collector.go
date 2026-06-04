package twitch

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	collectorName   = "twitch"
	metricNamespace = "billedapparat_collector_twitch_"
	bufferSize      = 1000
	avatarCacheSize = 100
)

type Collector struct {
	cfg               Config
	hubClient         *hubclient.HubClient
	twitchIrcClient   *twitch.Client
	twitchHelixClient *helix.Client
	avatarCache       *utils.LRUCache[string]
	eventsReceived    metric.Int64Counter
	eventsDropped     metric.Int64Counter
	logger            *slog.Logger
	msgBuffer         chan twitch.PrivateMessage
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("twitch-collector")

	eventsReceived, _ := meter.Int64Counter(metricNamespace+"messages_received_total",
		metric.WithDescription("Number of messages received from Twitch"))

	eventsDropped, _ := meter.Int64Counter(metricNamespace+"events_dropped_total",
		metric.WithDescription("Number of events dropped due to full buffer"))

	c := &Collector{
		cfg:               cfg,
		hubClient:         hubClient,
		avatarCache:       utils.NewLRUCache[string](avatarCacheSize),
		twitchIrcClient:   twitch.NewAnonymousClient(),
		twitchHelixClient: nil,
		eventsReceived:    eventsReceived,
		eventsDropped:     eventsDropped,
		logger:            slog.Default().With("component", "collector_twitch"),
		msgBuffer:         make(chan twitch.PrivateMessage, bufferSize),
	}

	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"worker_queue_depth",
		metric.WithDescription("Number of unprocessed events in the channel"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(len(c.msgBuffer)))

			return nil
		}),
	)

	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"avatar_cache_size",
		metric.WithDescription("Number of entries in the avatar cache"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(c.avatarCache.Stats().Size))

			return nil
		}),
	)

	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"avatar_cache_evictions",
		metric.WithDescription("Number of evictions in the avatar cache"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(c.avatarCache.Stats().Evictions))

			return nil
		}),
	)

	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"avatar_cache_hits",
		metric.WithDescription("Number of hits in the avatar cache"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(c.avatarCache.Stats().Hits))

			return nil
		}),
	)

	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"avatar_cache_misses",
		metric.WithDescription("Number of misses in the avatar cache"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(c.avatarCache.Stats().Misses))

			return nil
		}),
	)

	return c
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Run(ctx context.Context) error {
	if c.cfg.ClientID != "" && c.cfg.ClientSecret != "" {
		c.logger.Info("Authenticating Twitch Helix Client with provided credentials")

		helixClient, err := helix.NewClientWithContext(ctx, &helix.Options{
			ClientID:     c.cfg.ClientID,
			ClientSecret: c.cfg.ClientSecret,
		})
		if err != nil {
			return fmt.Errorf("failed to create Twitch Helix client: %w", err)
		}

		tokenResp, err := helixClient.RequestAppAccessToken([]string{})
		if err != nil {
			return fmt.Errorf("failed to request Twitch app access token: %w", err)
		}

		helixClient.SetAppAccessToken(tokenResp.Data.AccessToken)

		c.twitchHelixClient = helixClient
	}

	c.logger.Info("Starting Twitch Collector", "channel", c.cfg.Channel)

	c.twitchIrcClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
		c.handleMessageCreate(ctx, message)
	})

	c.twitchIrcClient.Join(c.cfg.Channel)

	go utils.RunWorker(ctx, c.msgBuffer, c.processRequest)

	go func() {
		<-ctx.Done()
		c.logger.Info("Shutting down Twitch IRC Client...")

		err := c.twitchIrcClient.Disconnect()
		if err != nil {
			c.logger.Error("Error during Twitch disconnect", "error", err)
		}
	}()

	err := c.twitchIrcClient.Connect()
	if err != nil {
		return fmt.Errorf("twitch IRC client stopped: %w", err)
	}

	c.logger.Info("Twitch Collector stopped gracefully")

	return nil
}

func (c *Collector) handleMessageCreate(ctx context.Context, m twitch.PrivateMessage) {
	if ctx.Err() != nil {
		return
	}

	c.logger.Debug("Received Twitch message", "user", m.User.DisplayName, "external_id", m.ID)
	c.eventsReceived.Add(ctx, 1)

	select {
	case c.msgBuffer <- m:
	default:
		c.logger.Warn("Message buffer full, dropping Twitch message", "user", m.User.DisplayName, "external_id", m.ID)
		c.eventsDropped.Add(ctx, 1)
	}
}
