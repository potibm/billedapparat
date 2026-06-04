package bluesky

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	collectorName    = "bluesky"
	jetstreamURL            = "wss://jetstream1.us-east.bsky.network/subscribe"
	metricNamespace         = "billedapparat_collector_bluesky_"
	eventBufferSize         = 1000
	profileRequestTimeout   = 5 * time.Second
	jetstreamReconnectDelay = 2 * time.Second
)

type Collector struct {
	cfg            Config
	searchTags     Hashtags
	hubClient      *hubclient.HubClient
	logger         *slog.Logger
	knownPosts     *RKeyList
	profiles       *ProfileList
	eventsChan     chan *JetstreamEvent
	eventsReceived metric.Int64Counter
	reconnects     metric.Int64Counter
	postsMatched   metric.Int64Counter
	wg             sync.WaitGroup
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("bluesky-collector")

	eventsReceived, _ := meter.Int64Counter(metricNamespace+"jetstream_events_received_total",
		metric.WithDescription("Number of events received from Jetstream"))
	reconnects, _ := meter.Int64Counter(metricNamespace+"jetstream_reconnects_total",
		metric.WithDescription("Number of connection drops/reconnects"))
	postsMatched, _ := meter.Int64Counter(metricNamespace+"jetstream_posts_matched_total",
		metric.WithDescription("Number of relevant posts (filtered)"))

	c := &Collector{
		cfg:            cfg,
		searchTags:     cfg.Hashtags.Normalize(),
		hubClient:      hubClient,
		logger:         slog.Default().With("component", "collector_bluesky"),
		knownPosts:     NewRKeyList(),
		profiles:       NewProfileList(),
		eventsChan:     make(chan *JetstreamEvent, eventBufferSize),
		eventsReceived: eventsReceived,
		reconnects:     reconnects,
		postsMatched:   postsMatched,
	}

	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"worker_queue_depth",
		metric.WithDescription("Number of unprocessed events in the channel"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(len(c.eventsChan)))

			return nil
		}),
	)
	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"profiles_cache_size",
		metric.WithDescription("Number of profiles currently cached"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(c.profiles.Len()))

			return nil
		}),
	)
	_, _ = meter.Int64ObservableGauge(
		metricNamespace+"known_posts_size",
		metric.WithDescription("Number of known posts currently tracked"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(c.knownPosts.Len()))

			return nil
		}),
	)

	return c
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Run(ctx context.Context) error {
	c.loadKnownPosts(ctx)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		c.processEvents(ctx)
	}()

	for ctx.Err() == nil {
		// before reconnecting, check if we should stop (e.g. STRG+C)
		err := c.connectAndRead(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.logger.Info("Received shutdown signal, stopping Collector...", "error", err)

				break
			}

			c.logger.Error("Lost connection, attempting to reconnect...", "error", err)
			c.reconnects.Add(ctx, 1)

			select {
			case <-time.After(jetstreamReconnectDelay):
			case <-ctx.Done():
				break
			}
		}
	}

	c.logger.Info("Closing events channel and waiting for worker to finish...")
	close(c.eventsChan)
	c.wg.Wait()
	c.logger.Info("Worker is finished. Program terminated safely.")

	return nil
}

func (c *Collector) connectAndRead(ctx context.Context) error {
	params := url.Values{}
	params.Add("wantedCollections", "app.bsky.feed.post")

	connectURL := fmt.Sprintf("%s?%s", jetstreamURL, params.Encode())

	conn, _, err := websocket.DefaultDialer.Dial(connectURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	readCtx, cancelRead := context.WithCancel(ctx)
	defer cancelRead()

	go func() {
		<-readCtx.Done()
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		c.eventsReceived.Add(ctx, 1)

		var event JetstreamEvent
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		if event.Kind == "commit" && event.Commit != nil && event.Commit.Collection == "app.bsky.feed.post" {
			c.eventsChan <- &event
		}
	}
}

func (c *Collector) loadKnownPosts(ctx context.Context) {
	c.logger.Info("Loading known posts from hub...")

	start := 0
	pageSize := 100

	for {
		externalIDs, total, err := c.hubClient.GetExternalIDs(ctx, collectorName, start, start+pageSize)
		if err != nil {
			c.logger.Error("Failed to fetch known posts from hub", "error", err)

			break
		}

		for _, id := range externalIDs {
			c.knownPosts.Add(id)
		}

		c.logger.Info("Fetched known posts page", "start", start, "count", len(externalIDs), "total", total)

		if start+pageSize >= total {
			break
		}

		start += pageSize
	}

	c.logger.Info("Known posts loading completed", "count", c.knownPosts.Len())
}

func (c *Collector) processEvents(ctx context.Context) {
	for event := range c.eventsChan {
		switch event.Commit.Operation {
		case "create":
			c.handleCreate(ctx, event)
		case "update":
			c.handleUpdate(ctx, event)
		case "delete":
			c.handleDelete(ctx, event)
		default:
			c.logger.Info("Unknown operation", "operation", event.Commit.Operation)
		}
	}
}
