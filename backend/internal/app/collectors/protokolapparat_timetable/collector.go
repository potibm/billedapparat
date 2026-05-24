package protokolapparat_timetable

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/protokolapparat/pkg/common"
	"github.com/potibm/protokolapparat/pkg/schedule"
	"github.com/redis/go-redis/v9"

	redisconsumer "github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_consumer"
)

const (
	protokolapparatTimetableScheduleCollector = "protokolapparat-timetable"
	protokolapparatVersion                    = 1
)

type Collector struct {
	hubClient *hubclient.HubClient
	logger    *slog.Logger
	consumer  *redisconsumer.Consumer[schedule.Entry]
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

	logger := slog.Default().With("component", "collector_protokolapparat_timetable")

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

func (c *Collector) pushToHub(ctx context.Context, event *common.Event[schedule.Entry]) error {
	switch event.Action {
	case common.ActionCreate, common.ActionUpdate:
		return c.handleCreateOrUpdate(ctx, event.Payload)
	case common.ActionSync:
		return c.handleSync(ctx, event.Payload)
	case common.ActionDelete:
		return c.handleDelete(ctx, event.Payload)
	default:
		return fmt.Errorf("unsupported action type: %s", event.Action)
	}
}

func (c *Collector) handleCreateOrUpdate(ctx context.Context, payloads []schedule.Entry) error {
	for _, payload := range payloads {
		ingestReq, err := mapPayload(payload)
		if err != nil {
			return fmt.Errorf("error mapping payload: %w", err)
		}

		if err := c.hubClient.SendTimetableEvent(ctx, ingestReq); err != nil {
			return fmt.Errorf("error sending timetable event to hub: %w", err)
		}
	}

	return nil
}

func (c *Collector) handleSync(ctx context.Context, payloads []schedule.Entry) error {
	var items []contracts.IngestTimetableEventRequest

	for _, payload := range payloads {
		ingestReq, err := mapPayload(payload)
		if err != nil {
			return fmt.Errorf("error mapping payload: %w", err)
		}

		items = append(items, ingestReq)
	}

	syncReq := contracts.IngestTimetableSyncRequest{
		Source: protokolapparatTimetableScheduleCollector,
		Items:  items,
	}

	c.logger.Info("Sending timetable sync to hub", "count", len(items))

	if err := c.hubClient.SendTimetableEventSync(ctx, syncReq); err != nil {
		return fmt.Errorf("error sending timetable sync to hub: %w", err)
	}

	return nil
}

func (c *Collector) handleDelete(ctx context.Context, payloads []schedule.Entry) error {
	for _, payload := range payloads {
		if err := c.hubClient.DeleteTimetableEvent(
			ctx,
			protokolapparatTimetableScheduleCollector,
			fmt.Sprintf("%d", payload.ID),
		); err != nil {
			return fmt.Errorf("error deleting timetable event from hub: %w", err)
		}
	}

	return nil
}

func mapPayload(payload schedule.Entry) (contracts.IngestTimetableEventRequest, error) {
	LocationName := ""
	LocationAddress := ""

	if payload.Location != nil {
		LocationName = payload.Location.Name
		LocationAddress = payload.Location.Address
	}

	CategoryName := ""
	CategoryColor := ""

	if payload.Category != nil {
		CategoryName = payload.Category.Name
		CategoryColor = payload.Category.Color
	}

	startTime, err := time.Parse(time.RFC3339, payload.StartTime)
	if err != nil {
		return contracts.IngestTimetableEventRequest{}, fmt.Errorf("invalid start_time %q: %w", payload.StartTime, err)
	}

	endTime, err := time.Parse(time.RFC3339, payload.EndTime)
	if err != nil && payload.EndTime != "" {
		return contracts.IngestTimetableEventRequest{}, fmt.Errorf("invalid end_time %q: %w", payload.EndTime, err)
	}

	return contracts.IngestTimetableEventRequest{
		Source:          protokolapparatTimetableScheduleCollector,
		ExternalID:      fmt.Sprintf("%d", payload.ID),
		Title:           payload.Title,
		Description:     payload.Description,
		StartTime:       startTime,
		EndTime:         endTime,
		LocationName:    LocationName,
		LocationAddress: LocationAddress,
		CategoryName:    CategoryName,
		CategoryColor:   CategoryColor,
		IsHidden:        payload.IsHidden,
	}, nil
}
