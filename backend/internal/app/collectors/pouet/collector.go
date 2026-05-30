package pouet

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

const (
	pouetCollectorName = "pouet"
	pouetOnelinerURL   = "https://www.pouet.net/oneliner.php"
)

type Collector struct {
	cfg       Config
	hubClient *hubclient.HubClient
	logger    *slog.Logger
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	return &Collector{
		cfg:       cfg,
		hubClient: hubClient,
		logger:    slog.Default().With("component", "collector_pouet"),
	}
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Run(ctx context.Context) error {
	interval := c.cfg.PollInterval

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	defer ticker.Stop()

	c.logger.Info("Starting Pouet Collector", "poll_interval_minutes", interval)

	if err := c.collectAndSend(ctx); err != nil {
		c.logger.Error("Error during initial collection", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := c.collectAndSend(ctx); err != nil {
				c.logger.Error("Error during collection", "error", err)
			}
		case <-ctx.Done():
			c.logger.Info("Pouet Collector stopping due to context cancellation")

			return nil
		}
	}
}

func (c *Collector) collectAndSend(ctx context.Context) error {
	c.logger.Info("Collecting data from Pouet")

	slideRequests, err := fetch(ctx, c.logger, pouetOnelinerURL)
	if err != nil {
		return fmt.Errorf("error fetching data: %w", err)
	}

	c.logger.Info("Fetched data from Pouet", "num_items", len(slideRequests))

	if len(c.cfg.Keywords) > 0 {
		slideRequests = filterByKeywords(slideRequests, c.cfg.Keywords.Lower())
		c.logger.Info("Filtered data by keywords", "num_items_after_filtering", len(slideRequests))
	}

	for _, req := range slideRequests {
		if err := c.hubClient.SendSlide(ctx, req); err != nil {
			c.logger.Error("Error ingesting slide to hub", "error", err, "external_id", req.ExternalID)

			continue
		}

		c.logger.Info("Successfully ingested slide to hub", "external_id", req.ExternalID)
	}

	return nil
}

func fetch(ctx context.Context, logger *slog.Logger, url string) ([]contracts.IngestSlideRequest, error) {
	// #nosec G107 -- url is passed as constant and not influenced by user input, so this is not vulnerable to SSRF
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating URL request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return parse(logger, resp.Body)
}
