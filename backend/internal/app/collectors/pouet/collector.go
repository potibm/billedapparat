package pouet

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/utils"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	collectorName    = "pouet"
	metricNamespace  = "billedapparat_collector_pouet_"
	pouetOnelinerURL = "https://www.pouet.net/oneliner.php"
	bufferSize       = 1000
	defaultTimeout   = 5 * time.Second
)

type Collector struct {
	cfg                Config
	httpClient         *http.Client
	hubClient          *hubclient.HubClient
	logger             *slog.Logger
	itemsFetchedTotal  metric.Int64Counter
	itemsFilteredTotal metric.Int64Counter
	msgBuffer          chan contracts.IngestSlideRequest
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("pouet-collector")

	itemsFetchedTotal, _ := meter.Int64Counter(metricNamespace+"items_fetched_total",
		metric.WithDescription("Number of items fetched from Pouet"))
	itemsFilteredTotal, _ := meter.Int64Counter(metricNamespace+"items_filtered_total",
		metric.WithDescription("Number of items filtered out due to keywords"))

	c := &Collector{
		cfg:                cfg,
		hubClient:          hubClient,
		logger:             slog.Default().With("component", "collector_pouet"),
		itemsFetchedTotal:  itemsFetchedTotal,
		itemsFilteredTotal: itemsFilteredTotal,
		msgBuffer:          make(chan contracts.IngestSlideRequest, bufferSize),
		httpClient:         &http.Client{Timeout: defaultTimeout},
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
	go utils.RunWorker(ctx, c.msgBuffer, c.processRequest)

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

	slideRequests, err := c.fetch(ctx, pouetOnelinerURL)
	if err != nil {
		return fmt.Errorf("error fetching data: %w", err)
	}

	c.logger.Info("Fetched data from Pouet", "num_items", len(slideRequests))
	c.itemsFetchedTotal.Add(ctx, int64(len(slideRequests)))

	if len(c.cfg.Keywords) > 0 {
		slideRequests = filterByKeywords(slideRequests, c.cfg.Keywords.Lower())
		c.logger.Info("Filtered data by keywords", "num_items_after_filtering", len(slideRequests))
		c.itemsFilteredTotal.Add(ctx, int64(len(slideRequests)))
	}

	for _, req := range slideRequests {
		c.msgBuffer <- req
	}

	return nil
}

func (c *Collector) fetch(ctx context.Context, url string) ([]contracts.IngestSlideRequest, error) {
	// #nosec G107 -- url is passed as constant and not influenced by user input, so this is not vulnerable to SSRF
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating URL request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return parse(c.logger, resp.Body)
}
