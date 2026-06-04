package discord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/utils"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	collectorName   = "discord"
	metricNamespace = "billedapparat_collector_discord_"
	bufferSize      = 1000
)

type Collector struct {
	cfg            Config
	hubClient      *hubclient.HubClient
	logger         *slog.Logger
	eventsReceived metric.Int64Counter
	eventsDropped  metric.Int64Counter
	postsMatched   metric.Int64Counter
	dg             *discordgo.Session
	msgBuffer      chan contracts.IngestSlideRequest
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("discord-collector")

	eventsReceived, _ := meter.Int64Counter(metricNamespace+"messages_received_total",
		metric.WithDescription("Number of messages received from Discord"))
	postsMatched, _ := meter.Int64Counter(metricNamespace+"messages_matched_total",
		metric.WithDescription("Number of relevant messages (filtered)"))
	eventsDropped, _ := meter.Int64Counter(metricNamespace+"messages_dropped_total",
		metric.WithDescription("Number of messages dropped due to full buffer"))

	c := &Collector{
		cfg:            cfg,
		hubClient:      hubClient,
		logger:         slog.Default().With("component", "collector_discord"),
		eventsReceived: eventsReceived,
		postsMatched:   postsMatched,
		eventsDropped:  eventsDropped,
		msgBuffer:      make(chan contracts.IngestSlideRequest, bufferSize),
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
	c.logger.Info("Shutting down Discord collector...")
	c.dg.Close()

	return nil
}

func (c *Collector) Run(ctx context.Context) error {
	var err error

	c.dg, err = discordgo.New("Bot " + c.cfg.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create Discord session: %w", err)
	}

	c.dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		c.handleMessageCreate(ctx, s, m)
	})

	c.dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	err = c.dg.Open()
	if err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	go utils.RunWorker(ctx, c.msgBuffer, c.processRequest)

	<-ctx.Done()

	return nil
}

func (c *Collector) handleMessageCreate(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	c.eventsReceived.Add(ctx, 1, metric.WithAttributes(
		attribute.String("mode", "create"),
	))

	if m.Author == nil {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.ChannelID != c.cfg.ChannelID {
		return
	}

	req := mapToIngestRequest(m.Message)

	c.postsMatched.Add(ctx, 1, metric.WithAttributes(
		attribute.String("mode", "create"),
	))

	select {
	case c.msgBuffer <- req:
	default:
		c.logger.Warn("Buffer overflow! Dropping Discord message", "msg_id", m.ID)
		c.eventsDropped.Add(ctx, 1)
	}
}
