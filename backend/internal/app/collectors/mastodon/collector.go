package mastodon

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	collectorName     = "mastodon"
	metricNamespace   = "billedapparat_collector_mastodon_"
	bufferSize        = 1000
	reconnectDuration = 5 * time.Second
)

type Collector struct {
	cfg              Config
	hubClient        *hubclient.HubClient
	logger           *slog.Logger
	reconnectCounter metric.Int64Counter
	eventsReceived   metric.Int64Counter
	msgBuffer        chan Event
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("mastodon-collector")

	eventsReceived, _ := meter.Int64Counter(metricNamespace+"messages_received_total",
		metric.WithDescription("Number of messages received from Mastodon"))

	reconnectCounter, _ := meter.Int64Counter(metricNamespace+"reconnects_total",
		metric.WithDescription("Number of reconnect attempts"))

	c := &Collector{
		cfg:              cfg,
		hubClient:        hubClient,
		logger:           slog.Default().With("component", "collector_mastodon"),
		eventsReceived:   eventsReceived,
		reconnectCounter: reconnectCounter,
		msgBuffer:        make(chan Event, bufferSize),
	}

	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"worker_queue_depth",
		metric.WithDescription("Number of unprocessed events in the channel"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(len(c.msgBuffer)))

			return nil
		}),
	)

	return c
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Run(ctx context.Context) error {
	if err := c.verifyCredentials(); err != nil {
		return err
	}

	streamURL := c.buildStreamingURL()
	c.logger.Info("Connect to Mastodon stream URL", "url", streamURL)

	go utils.RunWorker(ctx, c.msgBuffer, c.handleEvent)

	for {
		select {
		case <-ctx.Done(): // Strg+C
			c.logger.Info("Shutting down Mastodon collector")

			return nil
		default:
		}

		err := c.connectAndRead(ctx, streamURL)
		if err != nil {
			c.logger.Warn("Stream disconnected, reconnecting in 5s...", "error", err)
			c.reconnectCounter.Add(ctx, 1)

			timer := time.NewTimer(reconnectDuration)
			select {
			case <-ctx.Done():
				timer.Stop()

				return nil
			case <-timer.C:
			}
		}
	}
}

func (c *Collector) connectAndRead(ctx context.Context, streamURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, http.NoBody)
	if err != nil {
		return err
	}

	if c.cfg.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)
	}

	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)

	var currentEvent string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		c.eventsReceived.Add(ctx, 1)

		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "event:"):
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			if currentEvent != "" && payload != "" {
				event := Event{
					Type:    currentEvent,
					Payload: payload,
				}

				c.logger.Debug("Received event", "event", event)
				c.handleEvent(ctx, event)

				c.msgBuffer <- event
			}
		case line == "":
			currentEvent = ""
		}
	}
}

func (c *Collector) verifyCredentials() error {
	verifyURL := fmt.Sprintf("%s/api/v1/apps/verify_credentials", c.getBaseURL())

	req, err := http.NewRequest(http.MethodGet, verifyURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create verification request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error on access_token check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("mastodon access_token is invalid (401 Unauthorized)")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mastodon API reports an error: %s", resp.Status)
	}

	c.logger.Info("Successfully validated the Mastodon access_token!")

	return nil
}

func (c *Collector) getBaseURL() string {
	host := strings.TrimPrefix(c.cfg.Host, "https://")
	host = strings.TrimPrefix(host, "http://")

	return fmt.Sprintf("https://%s", host)
}

func (c *Collector) buildStreamingURL() string {
	baseURL := fmt.Sprintf("%s/api/v1/streaming", c.getBaseURL())

	if c.cfg.Tag != "" {
		return fmt.Sprintf("%s/hashtag?tag=%s", baseURL, url.QueryEscape(c.cfg.Tag))
	}

	return fmt.Sprintf("%s/public", baseURL)
}
