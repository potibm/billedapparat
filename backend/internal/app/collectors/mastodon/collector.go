package mastodon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
)

const mastodonCollectorName = "mastodon"

type Collector struct {
	cfg       Config
	hubClient *hubclient.HubClient
	logger    *slog.Logger
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	return &Collector{
		cfg:       cfg,
		hubClient: hubClient,
		logger:    slog.Default().With("component", "collector_mastodon"),
	}
}

func (c *Collector) Run(ctx context.Context) error {
	if err := c.verifyCredentials(); err != nil {
		return err
	}

	streamURL := c.buildStreamingURL()
	c.logger.Info("Connect to Mastodon stream URL", "url", streamURL)

	for {
		select {
		case <-ctx.Done(): // Strg+C
			c.logger.Info("Shutting down Mastodon collector")

			return nil
		default:
		}

		err := c.connectAndRead(ctx, streamURL)
		if err != nil {
			const reconnectDuration = 5 * time.Second

			c.logger.Warn("Stream disconnected, reconnecting in 5s...", "error", err)

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

		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "event:"):
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			if currentEvent != "" && payload != "" {
				c.handleEvent(currentEvent, payload)
			}
		case line == "":
			currentEvent = ""
		}
	}
}

func (c *Collector) handleEvent(eventType, payload string) {
	switch eventType {
	case "update", "status.update":
		var status MastoStatus

		c.logger.Debug("Received new status", "payload", payload)

		if err := json.Unmarshal([]byte(payload), &status); err != nil {
			c.logger.Error("Unable to parse post", "error", err)

			return
		}

		c.logger.Info("Received new post", "id", status.ID, "author", status.Account.Username)

		req := mapToIngestRequest(status)

		c.logger.Debug("Mapped post to ingest request", "request", req, "account", status.Account)

		if err := c.hubClient.SendSlide(req); err != nil {
			c.logger.Error("Error sending the post to the hub", "error", err)
		}

	case "delete":
		statusID := string(payload)
		c.logger.Info("Post was deleted", "id", statusID, "event_type", eventType)

		if err := c.hubClient.DeleteSlide(mastodonCollectorName, statusID); err != nil {
			c.logger.Error("Error sending delete request to the hub", "error", err)
		}
	default:
		c.logger.Info("Received unsupported event type", "type", eventType)
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
