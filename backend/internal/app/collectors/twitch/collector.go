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
	"go.opentelemetry.io/otel/attribute"
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
	logger            *slog.Logger
	msgBuffer         chan twitch.PrivateMessage
	metrics           utils.CollectorCounters
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("github.com/potibm/billedapparat/internal/app/collectors/twitch")

	c := &Collector{
		cfg:               cfg,
		hubClient:         hubClient,
		avatarCache:       utils.NewLRUCache[string](avatarCacheSize),
		twitchIrcClient:   twitch.NewAnonymousClient(),
		twitchHelixClient: nil,
		logger:            slog.Default().With("component", "collector_twitch"),
		msgBuffer:         make(chan twitch.PrivateMessage, bufferSize),
		metrics:           utils.NewCollectorCounters(meter),
	}

	utils.RegisterQueueDepthGauge(meter, collectorName, func() int {
		return len(c.msgBuffer)
	})

	utils.RegisterCacheMetrics(meter, collectorName, "avatar_cache", c.avatarCache)

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
	c.metrics.EventsReceived.Add(ctx, 1, metric.WithAttributes(
		attribute.String("collector", collectorName),
	))

	select {
	case c.msgBuffer <- m:
	default:
		c.logger.Warn("Message buffer full, dropping Twitch message", "user", m.User.DisplayName, "external_id", m.ID)
		c.metrics.EventsDropped.Add(ctx, 1, metric.WithAttributes(
			attribute.String("collector", collectorName),
		))
	}
}
