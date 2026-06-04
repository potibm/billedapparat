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
	collectorName = "discord"
	bufferSize    = 1000
)

type Collector struct {
	cfg       Config
	hubClient *hubclient.HubClient
	logger    *slog.Logger
	dg        *discordgo.Session
	msgBuffer chan contracts.IngestSlideRequest
	metrics   utils.CollectorCounters
}

func NewCollector(cfg Config, hubClient *hubclient.HubClient) *Collector {
	meter := otel.Meter("github.com/potibm/billedapparat/internal/app/collectors/discord")

	c := &Collector{
		cfg:       cfg,
		hubClient: hubClient,
		logger:    slog.Default().With("component", "collector_discord"),
		msgBuffer: make(chan contracts.IngestSlideRequest, bufferSize),
		metrics:   utils.NewCollectorCounters(meter),
	}

	utils.RegisterQueueDepthGauge(meter, collectorName, func() int {
		return len(c.msgBuffer)
	})

	return c
}

func (c *Collector) Close() error {
	if c.dg != nil {
		c.logger.Info("Shutting down Discord collector...")
		c.dg.Close()
	}

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
	c.metrics.EventsReceived.Add(ctx, 1, metric.WithAttributes(
		attribute.String("collector", collectorName),
		attribute.String("operation", "create"),
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

	c.metrics.EventsMatched.Add(ctx, 1, metric.WithAttributes(
		attribute.String("collector", collectorName),
		attribute.String("operation", "create"),
	))

	select {
	case c.msgBuffer <- req:
	default:
		c.logger.Warn("Buffer overflow! Dropping Discord message", "msg_id", m.ID)
		c.metrics.EventsDropped.Add(ctx, 1, metric.WithAttributes(
			attribute.String("collector", collectorName),
		))
	}
}
