package redisconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/potibm/protokolapparat/pkg/common"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	StreamName    string
	ConsumerGroup string
	ConsumerName  string
}

type Consumer[T common.Validatable] struct {
	cfg         Config
	rdb         *redis.Client
	logger      *slog.Logger
	version     int
	pushHandler func(ctx context.Context, event *common.Event[T]) error
}

func New[T common.Validatable](
	cfg Config,
	rdb *redis.Client,
	logger *slog.Logger,
	version int,
	pushHandler func(ctx context.Context, event *common.Event[T]) error,
) *Consumer[T] {
	return &Consumer[T]{
		cfg:         cfg,
		rdb:         rdb,
		logger:      logger,
		version:     version,
		pushHandler: pushHandler,
	}
}

func (c *Consumer[T]) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}

	return nil
}

func (c *Consumer[T]) Run(ctx context.Context) error {
	const (
		batchSize   = 50
		pollTimeout = 2 * time.Second
	)

	startID := "0"

	if err := c.ensureConsumerGroup(ctx); err != nil {
		return err
	}

	c.logger.Info("Started redis collector", "stream", c.cfg.StreamName, "group", c.cfg.ConsumerGroup)

	for {
		if ctx.Err() != nil {
			c.logger.Info("Collector terminates...")

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
				c.logger.Debug("PEL is empty, starting with new messages (>)")

				startID = ">"
			}

			continue
		}

		c.processStreams(ctx, streams)
	}
}

func (c *Consumer[T]) ensureConsumerGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("error while creating consumer group: %w", err)
	}

	return nil
}

func (c *Consumer[T]) handleReadError(err error) bool {
	if errors.Is(err, redis.Nil) {
		return false // timeout
	}

	if errors.Is(err, context.Canceled) {
		return true // Graceful Shutdown
	}

	c.logger.Error("Error reading from Redis Stream", "err", err)
	time.Sleep(1 * time.Second) // spam protection

	return false
}

func (c *Consumer[T]) processStreams(ctx context.Context, streams []redis.XStream) {
	for _, stream := range streams {
		for _, message := range stream.Messages {
			c.processSingleMessage(ctx, message)
		}
	}
}

func (c *Consumer[T]) processSingleMessage(ctx context.Context, message redis.XMessage) {
	event, err := c.parseRedisMessage(message)
	if err != nil {
		c.logger.Error("Could not parse message (Poison Pill) - discarding it", "id", message.ID, "err", err)
		c.rdb.XAck(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, message.ID)

		return
	}

	c.logger.Debug("Received message", "id", message.ID, "action", event.Action)

	if err := c.pushHandler(ctx, event); err != nil {
		c.logger.Error("Error processing message, will retry later", "id", message.ID, "err", err)

		return
	}

	c.rdb.XAck(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, message.ID)
	c.logger.Debug("Message processed", "id", message.ID)
}

func (c *Consumer[T]) parseRedisMessage(message redis.XMessage) (*common.Event[T], error) {
	rawStr, ok := message.Values["data"].(string)
	if !ok {
		return nil, fmt.Errorf("expected key 'data' missing or not a string")
	}

	var event common.Event[T]
	if err := json.Unmarshal([]byte(rawStr), &event); err != nil {
		return nil, fmt.Errorf("error parsing message ID %s: %w", message.ID, err)
	}

	if event.Version != c.version {
		return nil, fmt.Errorf("unsupported schema version: %d", event.Version)
	}

	return &event, nil
}
