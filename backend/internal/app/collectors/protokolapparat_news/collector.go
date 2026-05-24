package protokolapparat_news

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/protokolapparat/pkg/common"
	"github.com/potibm/protokolapparat/pkg/news"
	"github.com/redis/go-redis/v9"

	redisconsumer "github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_consumer"
)

const (
	protokolapparatNewsCollector = "protokolapparat-news"
	protokolapparatVersion       = 1
)

type Collector struct {
	hubClient *hubclient.HubClient
	logger    *slog.Logger
	consumer  *redisconsumer.Consumer[news.Entry] // Der generische Consumer
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

	logger := slog.Default().With("component", "collector_protokolapparat_news")

	c := &Collector{
		hubClient: hubClient,
		logger:    logger,
	}

	c.consumer = redisconsumer.New(
		redisconsumer.Config{
			StreamName:    cfg.StreamName,
			ConsumerGroup: cfg.ConsumerGroup,
			ConsumerName:  cfg.ConsumerName,
		},
		rdb,
		logger,
		protokolapparatVersion,
		c.pushToHub,
	)

	return c
}

func (c *Collector) Close() error {
	return c.consumer.Close()
}

func (c *Collector) Run(ctx context.Context) error {
	return c.consumer.Run(ctx)
}

func (c *Collector) pushToHub(ctx context.Context, event *common.Event[news.Entry]) error {
	switch event.Action {
	case common.ActionCreate, common.ActionUpdate:
		for _, payload := range event.Payload {
			ingestReq := mapPayload(payload)

			if err := c.hubClient.SendNews(ctx, ingestReq); err != nil {
				return fmt.Errorf("error sending news to hub: %w", err)
			}
		}

	case common.ActionSync:
		var items []contracts.IngestNewsRequest

		for _, payload := range event.Payload {
			items = append(items, mapPayload(payload))
		}

		syncReq := contracts.IngestNewsSyncRequest{
			Source: protokolapparatNewsCollector,
			Items:  items,
		}

		c.logger.Info("Sending news sync to hub", "count", len(items))

		if err := c.hubClient.SendNewsSync(ctx, syncReq); err != nil {
			return fmt.Errorf("error sending news sync to hub: %w", err)
		}

	case common.ActionDelete:
		for _, payload := range event.Payload {
			if err := c.hubClient.DeleteNews(
				ctx,
				protokolapparatNewsCollector,
				fmt.Sprintf("%d", payload.ID),
			); err != nil {
				return fmt.Errorf("error deleting news from hub: %w", err)
			}
		}

	default:
		return fmt.Errorf("unsupported action type: %s", event.Action)
	}

	return nil
}

func mapPayload(payload news.Entry) contracts.IngestNewsRequest {
	return contracts.IngestNewsRequest{
		Source:      protokolapparatNewsCollector,
		ExternalID:  fmt.Sprintf("%d", payload.ID),
		Title:       payload.Title,
		Body:        payload.Body,
		IsUrgent:    payload.IsUrgent,
		ExternalURL: payload.ExternalURL,
		IsHidden:    payload.IsHidden,
	}
}
