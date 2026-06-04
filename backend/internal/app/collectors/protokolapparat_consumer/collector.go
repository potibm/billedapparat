package redisconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/utils"
	"github.com/potibm/protokolapparat/pkg/common"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	pelStartID    = "0"
	newMessagesID = ">"
)

type Config struct {
	StreamName    string
	ConsumerGroup string
	ConsumerName  string
}

type Consumer[T common.Validatable] struct {
	cfg            Config
	rdb            *redis.Client
	logger         *slog.Logger
	version        int
	pushHandler    func(ctx context.Context, event *common.Event[T]) error
	pushErrorCount int
	collectorName  string
	metrics        utils.CollectorCounters
}

func New[T common.Validatable](
	cfg Config,
	rdb *redis.Client,
	logger *slog.Logger,
	version int,
	collectorName string,
	pushHandler func(ctx context.Context, event *common.Event[T]) error,
) *Consumer[T] {
	meter := otel.Meter("github.com/potibm/billedapparat/internal/app/collectors/redisconsumer")

	return &Consumer[T]{
		cfg:            cfg,
		rdb:            rdb,
		logger:         logger,
		version:        version,
		pushHandler:    pushHandler,
		pushErrorCount: 0,
		collectorName:  collectorName,
		metrics:        utils.NewCollectorCounters(meter),
	}
}

func (c *Consumer[T]) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}

	return nil
}

func (c *Consumer[T]) Run(ctx context.Context) error {
	if err := c.ensureConsumerGroup(ctx); err != nil {
		return err
	}

	c.logger.Info("Started redis collector", "stream", c.cfg.StreamName, "group", c.cfg.ConsumerGroup)

	startID := pelStartID // Always start by clearing the Pending Entries List (PEL)

	// Keep looping as long as the context is active
	for ctx.Err() == nil {
		var terminate bool

		startID, terminate = c.processNextBatch(ctx, startID)
		if terminate {
			break
		}
	}

	c.logger.Info("Collector terminates...")

	return nil
}

func (c *Consumer[T]) processNextBatch(ctx context.Context, startID string) (nextID string, terminate bool) {
	const (
		batchSize   = 50
		pollTimeout = 2 * time.Second
	)

	args := &redis.XReadGroupArgs{
		Group:    c.cfg.ConsumerGroup,
		Consumer: c.cfg.ConsumerName,
		Streams:  []string{c.cfg.StreamName, startID},
		Count:    batchSize,
		Block:    pollTimeout,
	}

	streams, err := c.rdb.XReadGroup(ctx, args).Result()
	if err != nil {
		return startID, c.handleReadError(err)
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		if startID == pelStartID {
			c.logger.Debug("PEL is empty, starting with new messages (>)")

			return newMessagesID, false
		}

		return startID, false
	}

	hadFailures := c.processStreams(ctx, streams)

	return c.handleBatchResult(ctx, startID, hadFailures)
}

func (c *Consumer[T]) handleBatchResult(
	ctx context.Context,
	startID string,
	hadFailures bool,
) (nextID string, terminate bool) {
	// If everything succeeded, we continue
	if !hadFailures {
		return startID, false
	}

	nextID = startID
	if nextID != pelStartID {
		c.logger.Info("Message processing failed, falling back to PEL (0) for retries")

		nextID = pelStartID
	}

	backoffDelay := calculateBackoff(c.pushErrorCount)
	c.logger.Info("Pausing message consumption due to errors", "backoff", backoffDelay, "errors", c.pushErrorCount)

	// Context-aware backoff to prevent a tight loop on persistently failing messages
	select {
	case <-ctx.Done():
		return pelStartID, true
	case <-time.After(backoffDelay):
		c.metrics.Reconnects.Add(ctx, 1, metric.WithAttributes(
			attribute.String("collector", c.collectorName),
		))

		return nextID, false
	}
}

func (c *Consumer[T]) ensureConsumerGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, pelStartID).Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("error while creating consumer group: %w", err)
	}

	return nil
}

func calculateBackoff(errorCount int) time.Duration {
	const (
		baseDelay = 1 * time.Second
		maxDelay  = 5 * time.Minute
		maxShifts = 30
	)

	if errorCount <= 1 {
		return baseDelay
	}

	shifts := errorCount - 1

	if shifts > maxShifts {
		return maxDelay
	}

	multiplier := 1 << shifts

	delay := baseDelay * time.Duration(multiplier)

	if delay > maxDelay {
		return maxDelay
	}

	return delay
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

// processStreams now returns true if ANY message in the batch failed to process or ACK.
func (c *Consumer[T]) processStreams(ctx context.Context, streams []redis.XStream) bool {
	hadFailures := false

	for _, stream := range streams {
		for _, message := range stream.Messages {
			success := c.processSingleMessage(ctx, message)
			if !success {
				hadFailures = true
			}
		}
	}

	return hadFailures
}

// processSingleMessage now returns a boolean indicating success (true) or failure (false).
func (c *Consumer[T]) processSingleMessage(ctx context.Context, message redis.XMessage) bool {
	c.metrics.EventsMatched.Add(ctx, 1, metric.WithAttributes(
		attribute.String("collector", c.collectorName),
	))

	event, err := c.parseRedisMessage(message)
	if err != nil {
		c.logger.Error("Could not parse message (Poison Pill) - discarding it", "id", message.ID, "err", err)

		// Added .Err() check for XAck
		if ackErr := c.rdb.XAck(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, message.ID).Err(); ackErr != nil {
			c.logger.Error("Failed to ACK poison pill message", "id", message.ID, "err", ackErr)
		}

		// Return true because a poison pill is intentionally discarded, so we don't want to retry it.
		return true
	}

	c.logger.Debug("Received message", "id", message.ID, "action", event.Action)

	if err := c.pushHandler(ctx, event); err != nil {
		c.pushErrorCount++
		c.logger.Error(
			"Error processing message, will retry later",
			"id",
			message.ID,
			"err",
			err,
			"consecutive_errors",
			c.pushErrorCount,
		)

		return false // Processing failed, trigger PEL retry
	} else {
		c.pushErrorCount = 0
	}

	// Added .Err() check for XAck
	if ackErr := c.rdb.XAck(ctx, c.cfg.StreamName, c.cfg.ConsumerGroup, message.ID).Err(); ackErr != nil {
		c.logger.Error("Failed to ACK processed message", "id", message.ID, "err", ackErr)
		// Assuming at-least-once delivery semantics (idempotency), an ACK failure should trigger a retry.
		return false
	}

	c.logger.Debug("Message processed", "id", message.ID)

	return true
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
