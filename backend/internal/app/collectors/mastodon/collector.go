package mastodon

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	collectorName     = "mastodon"
	bufferSize        = 1000
	reconnectDuration = 5 * time.Second
)

type Collector struct {
	cfg        Config
	httpClient *http.Client
	hubClient  *hubclient.HubClient
	logger     *slog.Logger
	msgBuffer  chan Event
	metrics    utils.CollectorCounters
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("github.com/potibm/billedapparat/internal/app/collectors/mastodon")

	c := &Collector{
		cfg:        cfg,
		httpClient: &http.Client{},
		hubClient:  hubClient,
		logger:     slog.Default().With("component", "collector_mastodon"),
		msgBuffer:  make(chan Event, bufferSize),
		metrics:    utils.NewCollectorCounters(meter),
	}

	utils.RegisterQueueDepthGauge(meter, collectorName, func() int {
		return len(c.msgBuffer)
	})

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

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		utils.RunWorker(ctx, c.msgBuffer, c.handleEvent)
	}()

	defer func() {
		c.logger.Info("Shutting down Mastodon collector: closing buffer and draining events")
		close(c.msgBuffer)
		wg.Wait()
		c.logger.Info("Mastodon collector shutdown complete")
	}()

	for {
		if ctx.Err() != nil {
			return nil
		}

		err := c.connectAndRead(ctx, streamURL)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			c.logger.Warn("Stream disconnected, reconnecting in 5s...", "error", err)
			c.metrics.Reconnects.Add(ctx, 1, metric.WithAttributes(
				attribute.String("collector", collectorName),
			))

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
			c.metrics.EventsReceived.Add(ctx, 1, metric.WithAttributes(
				attribute.String("collector", collectorName),
			))

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
		c.metrics.EventsDropped.Add(ctx, 1, metric.WithAttributes(
			attribute.String("collector", collectorName),
		))
	}
}

func (c *Collector) verifyCredentials(ctx context.Context) error {
	const defaultTimeout = 5 * time.Second

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

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
