package protokolapparat_news

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/protokolapparat/pkg/common"
	"github.com/potibm/protokolapparat/pkg/news"
	"github.com/redis/go-redis/v9"
)

const (
	protokolapparatNewsCollector = "protokolapparat-news"
	protokolapparatVersion       = 1
)

type Collector struct {
	cfg       Config
	rdb       *redis.Client
	hubClient *hubclient.HubClient
	logger    *slog.Logger
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient, rdb *redis.Client) *Collector {
	if cfg.StreamName == "" {
		cfg.StreamName = defaultStreamName
	}

	if cfg.ConsumerGroup == "" {
		cfg.ConsumerGroup = defaultConsumerGroup
	}

	if cfg.ConsumerName == "" {
		cfg.ConsumerName = defaultConsumerName
	}

	return &Collector{
		cfg:       cfg,
		hubClient: hubClient,
		rdb:       rdb,
		logger:    slog.Default().With("component", "collector_protokolapparat_news"),
	}
}

func (c *Collector) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}

	return nil
}

func (c *Collector) Run(ctx context.Context) error {
	const (
		batchSize   = 50
		pollTimeout = 2 * time.Second
	)

	startID := "0"

	if err := c.ensureConsumerGroup(ctx); err != nil {
		return err
	}

	slog.Info("Started redis collector", "stream", c.cfg.StreamName, "group", c.cfg.ConsumerGroup)

	for {
		if ctx.Err() != nil {
			slog.Info("Collector terminates...")

			return nil
		}

		args := &redis.XReadGroupArgs{
			Group:    c.cfg.ConsumerGroup,
			Consumer: c.cfg.ConsumerName,
			Streams:  []string{c.cfg.StreamName, startID},
			Count:    batchSize,
			Block:    pollTimeout,
		}

		streams, err := c.rdb.XReadGroup(ctx, args).Result()
		if err != nil {
			if c.handleReadError(err) {
				return nil
			}

			continue
		}

		if len(streams) == 0 || len(streams[0].Messages) == 0 {
			if startID == "0" {
				slog.Debug("PEL is empty, starting with new messages (>)")

				startID = ">"
			}

			continue
		}

		c.processStreams(ctx, streams)
	}
}

func (c *Collector) ensureConsumerGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("error while creating consumer group: %w", err)
	}

	return nil
}

func (c *Collector) handleReadError(err error) bool {
	if errors.Is(err, redis.Nil) {
		return false // timeout
	}

	if errors.Is(err, context.Canceled) {
		return true // Graceful Shutdown
	}

	slog.Error("Error reading from Redis Stream", "err", err)
	time.Sleep(1 * time.Second) // spam protection

	return false
}

func (c *Collector) processStreams(ctx context.Context, streams []redis.XStream) {
	for _, stream := range streams {
		for _, message := range stream.Messages {
			c.processSingleMessage(ctx, message)
		}
	}
}

func (c *Collector) processSingleMessage(ctx context.Context, message redis.XMessage) {
	newsEvent, err := c.parseRedisMessage(message)
	if err != nil {
		slog.Error("Could not parse message (Poison Pill) - discarding it", "id", message.ID, "err", err)
		c.rdb.XAck(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, message.ID)

		return
	}

	slog.Debug("Received message", "id", message.ID, "action", newsEvent.Action)

	if err := c.pushToHub(ctx, newsEvent); err != nil {
		slog.Error("Error sending to hub, will retry later", "id", message.ID, "err", err)

		return
	}

	c.rdb.XAck(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, message.ID)
	slog.Debug("Message processed", "id", message.ID)
}

func (c *Collector) pushToHub(ctx context.Context, event *common.Event[news.Entry]) error {
	switch event.Action {
	case common.ActionCreate, common.ActionUpdate:
		for _, payload := range event.Payload {
			ingestReq := contracts.IngestNewsRequest{
				Source:      protokolapparatNewsCollector,
				ExternalID:  fmt.Sprintf("%d", payload.ID),
				Title:       payload.Title,
				Body:        payload.Body,
				IsUrgent:    payload.IsUrgent,
				ExternalURL: payload.ExternalURL,
				IsHidden:    payload.IsHidden,
			}

			if err := c.hubClient.SendNews(ingestReq); err != nil {
				return fmt.Errorf("error sending news to hub: %w", err)
			}
		}

	case common.ActionSync:
		var items []contracts.IngestNewsRequest

		for _, payload := range event.Payload {
			items = append(items, contracts.IngestNewsRequest{
				Source:      protokolapparatNewsCollector,
				ExternalID:  fmt.Sprintf("%d", payload.ID),
				Title:       payload.Title,
				Body:        payload.Body,
				IsUrgent:    payload.IsUrgent,
				ExternalURL: payload.ExternalURL,
				IsHidden:    payload.IsHidden,
			})
		}

		syncReq := contracts.IngestNewsSyncRequest{
			Source: protokolapparatNewsCollector,
			Items:  items,
		}

		c.logger.Info("Sending news sync to hub", "count", len(items))

		if err := c.hubClient.SendNewsSync(syncReq); err != nil {
			return fmt.Errorf("error sending news sync to hub: %w", err)
		}

	case common.ActionDelete:
		for _, payload := range event.Payload {
			if err := c.hubClient.DeleteNews(protokolapparatNewsCollector, fmt.Sprintf("%d", payload.ID)); err != nil {
				return fmt.Errorf("error deleting news from hub: %w", err)
			}
		}

	default:
		return fmt.Errorf("unsupported action type: %s", event.Action)
	}

	return nil
}

func (c *Collector) parseRedisMessage(message redis.XMessage) (*common.Event[news.Entry], error) {
	rawStr, ok := message.Values["data"].(string)
	if !ok {
		return nil, fmt.Errorf("expected key 'data' missing or not a string")
	}

	var event common.Event[news.Entry]

	if err := json.Unmarshal([]byte(rawStr), &event); err != nil {
		return nil, fmt.Errorf("error parsing message ID %s: %w", message.ID, err)
	}

	if event.Version != protokolapparatVersion {
		return nil, fmt.Errorf("unsupported schema version: %d", event.Version)
	}

	return &event, nil
}
