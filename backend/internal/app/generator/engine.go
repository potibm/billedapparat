package generator

import (
	"context"
	"log/slog"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
)

type SlideGenerator interface {
	Name() string

	Generate(ctx context.Context) ([]domain.Slide, error)
}

type SSEBroadcaster interface {
	Broadcast(event domain.StreamEvent, payload any)
}

type Engine struct {
	generators     []SlideGenerator
	slideRepo      repository.SlideRepository
	sseBroadcaster SSEBroadcaster
	interval       time.Duration
	logger         *slog.Logger
	triggerCh      chan string
}

func NewEngine(
	repo repository.SlideRepository,
	sseBroadcaster SSEBroadcaster,
	logger *slog.Logger,
	gens ...SlideGenerator,
) *Engine {
	return &Engine{
		generators:     gens,
		slideRepo:      repo,
		sseBroadcaster: sseBroadcaster,
		interval:       1 * time.Minute,
		triggerCh:      make(chan string, 1),
		logger:         logger,
	}
}

func (e *Engine) Trigger(source string) {
	e.logger.Info("Triggering generator", "source", source)

	select {
	case e.triggerCh <- source:
		// signal sent successfully
		e.logger.Debug("Trigger signal queued successfully", "source", source)
	default:
		// If the channel is full, it means a trigger is already pending, so we can skip this one.
		e.logger.Warn("Trigger channel full or nil, skipping", "source", source)
	}
}

func (e *Engine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	e.runGenerators(ctx, "all")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			e.runGenerators(ctx, "all")
		case source := <-e.triggerCh:
			e.runGenerators(ctx, source)
		}
	}
}

func (e *Engine) runGenerators(ctx context.Context, requestedSource string) {
	e.logger.Info("Running generators", "source", requestedSource)

	for _, gen := range e.generators {
		if requestedSource != "all" && gen.Name() != requestedSource {
			continue
		}

		slides, err := gen.Generate(ctx)
		if err != nil {
			e.logger.Error("Generator failed", "name", gen.Name(), "err", err)

			continue
		}

		result, err := e.slideRepo.Sync(ctx, gen.Name(), slides)
		if err != nil {
			e.logger.Error("Failed to sync slides to DB", "err", err)

			continue
		}

		for _, s := range result.Created {
			e.sseBroadcaster.Broadcast(domain.EventCreate, s)
		}

		for _, s := range result.Updated {
			e.sseBroadcaster.Broadcast(domain.EventUpdate, s)
		}

		for _, s := range result.Deleted {
			e.sseBroadcaster.Broadcast(domain.EventDelete, s.ID)
		}
	}
}
