package lorem

import (
	"context"
	"log/slog"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/utils"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	collectorName = "lorem"
)

type Collector struct {
	cfg       Config
	hubClient *hubclient.HubClient
	logger    *slog.Logger
	metrics   utils.CollectorCounters
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("github.com/potibm/billedapparat/internal/app/collectors/lorem")

	return &Collector{
		cfg:       cfg,
		hubClient: hubClient,
		logger:    slog.Default().With("component", "collector_lorem"),
		metrics:   utils.NewCollectorCounters(meter),
	}
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Run(ctx context.Context) error {
	interval := c.cfg.PollInterval
	if interval <= 0 {
		interval = loremDefaultPollIntervalSeconds
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	c.logger.Info("Starting Lorem Collector", "poll_interval_seconds", interval)

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
			c.logger.Info("Lorem Collector stopping due to context cancellation")

			return nil
		}
	}
}

func (c *Collector) collectAndSend(ctx context.Context) error {
	c.logger.Info("Generating dummy data for Lorem")

	slideRequests := c.generateDummySlides()

	c.logger.Info("Generated dummy data", "num_items", len(slideRequests))
	c.metrics.EventsReceived.Add(ctx, int64(len(slideRequests)), metric.WithAttributes(
		attribute.String("collector", collectorName),
	))

	c.metrics.EventsMatched.Add(ctx, int64(len(slideRequests)), metric.WithAttributes(
		attribute.String("collector", collectorName),
	))

	for _, req := range slideRequests {
		err := c.hubClient.SendSlide(ctx, req)
		if err != nil {
			c.logger.Error("Failed to send slide to hub", "external_id", req.ExternalID, "error", err)
			c.metrics.EventsDropped.Add(ctx, 1, metric.WithAttributes(
				attribute.String("collector", collectorName),
			))

			continue
		}

		c.logger.Info("Successfully sent dummy slide", "external_id", req.ExternalID)
	}

	return nil
}

func (c *Collector) generateDummySlides() []contracts.IngestSlideRequest {
	now := time.Now()
	bodyLength := 20

	req := contracts.IngestSlideRequest{
		Source:     collectorName,
		ExternalID: gofakeit.UUID(),
		Author: &contracts.IngestSlideRequestAuthor{
			ExternalID:  gofakeit.UUID(),
			Username:    gofakeit.Username(),
			DisplayName: gofakeit.Name(),
		},
		Body:            gofakeit.Sentence(bodyLength),
		Language:        "en",
		OriginCreatedAt: now,
	}

	return []contracts.IngestSlideRequest{req}
}
