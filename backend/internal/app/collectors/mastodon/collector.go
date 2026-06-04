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
	defaultTimeout    = 5 * time.Second
)

type Collector struct {
	cfg              Config
	httpClient       *http.Client
	hubClient        *hubclient.HubClient
	logger           *slog.Logger
	reconnectCounter metric.Int64Counter
	eventsReceived   metric.Int64Counter
	eventsDropped    metric.Int64Counter
	msgBuffer        chan Event
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("mastodon-collector")

	eventsReceived, _ := meter.Int64Counter(metricNamespace+"messages_received_total",
		metric.WithDescription("Number of messages received from Mastodon"))

	reconnectCounter, _ := meter.Int64Counter(metricNamespace+"reconnects_total",
		metric.WithDescription("Number of reconnect attempts"))

	eventsDropped, _ := meter.Int64Counter(metricNamespace+"messages_dropped_total",
		metric.WithDescription("Number of messages dropped due to full buffer"))

	c := &Collector{
		cfg:              cfg,
		httpClient:       &http.Client{Timeout: defaultTimeout},
		hubClient:        hubClient,
		logger:           slog.Default().With("component", "collector_mastodon"),
		eventsReceived:   eventsReceived,
		reconnectCounter: reconnectCounter,
		eventsDropped:    eventsDropped,
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
	if err := c.verifyCredentials(ctx); err != nil {
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

	resp, err := c.httpClient.Do(req)
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

		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "event:"):
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			c.eventsReceived.Add(ctx, 1)

			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			c.dispatchEvent(ctx, currentEvent, payload)
		case line == "":
			currentEvent = ""
		}
	}
}

func (c *Collector) dispatchEvent(ctx context.Context, eventType, payload string) {
	if eventType == "" || payload == "" {
		return
	}

	event := Event{
		Type:    eventType,
		Payload: payload,
	}

	c.logger.Debug("Received event", "event", event)

	select {
	case c.msgBuffer <- event:
	default:
		c.logger.Warn("Message buffer full, dropping Mastodon event", "event_type", event.Type)
		c.eventsDropped.Add(ctx, 1)
	}
}

func (c *Collector) verifyCredentials(ctx context.Context) error {
	verifyURL := fmt.Sprintf("%s/api/v1/apps/verify_credentials", c.getBaseURL())
	opts := utils.RequestOptions{
		Headers: map[string]string{
			"Authorization": "Bearer " + c.cfg.AccessToken,
		},
	}

	resp, err := utils.DoGet(ctx, c.httpClient, verifyURL, opts)
	if err != nil {
		return err
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
