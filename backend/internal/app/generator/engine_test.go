package generator

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGenerator implements SlideGenerator for testing.
type mockGenerator struct {
	name   string
	slides []domain.Slide
	err    error
}

func (m *mockGenerator) Name() string { return m.name }

func (m *mockGenerator) Generate(ctx context.Context) ([]domain.Slide, error) {
	return m.slides, m.err
}

// mockBroadcaster implements SSEBroadcaster for testing.
type mockBroadcaster struct {
	events   []domain.StreamEvent
	payloads []any
	mu       sync.Mutex
}

func (m *mockBroadcaster) Broadcast(event domain.StreamEvent, payload any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = append(m.events, event)
	m.payloads = append(m.payloads, payload)
}

// mockSlideRepo implements only Sync for testing.
type mockSlideRepo struct {
	result *repository.SlideSyncResult
	err    error
}

func (m *mockSlideRepo) Sync(
	ctx context.Context,
	source string,
	slides []domain.Slide,
) (*repository.SlideSyncResult, error) {
	return m.result, m.err
}

// Stub methods to satisfy repository.SlideRepository interface.
func (m *mockSlideRepo) Save(ctx context.Context, slide *domain.Slide) error { return nil }

func (m *mockSlideRepo) GetActive(ctx context.Context) ([]domain.Slide, error) { return nil, nil }

func (m *mockSlideRepo) Delete(ctx context.Context, id int64) error { return nil }

func (m *mockSlideRepo) AdminList(
	ctx context.Context,
	params repository.SlideListParams,
	filters repository.SlideListFilters,
) ([]domain.Slide, int64, error) {
	return nil, 0, nil
}

func (m *mockSlideRepo) GetByID(ctx context.Context, id int64) (*domain.Slide, error) {
	return nil, nil
}

func (m *mockSlideRepo) MarkAsDeleted(ctx context.Context, source, externalID string) error {
	return nil
}

func (m *mockSlideRepo) SlideExists(source, externalID string, subID *int) (bool, error) {
	return false, nil
}

func (m *mockSlideRepo) GetAllMediaURLs(ctx context.Context) ([]string, error) { return nil, nil }

func (m *mockSlideRepo) FindLocalURLByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	return "", false
}

func (m *mockSlideRepo) FindExpiredSlidesByType(
	ctx context.Context,
	slideType string,
	cutoff time.Time,
) ([]domain.Slide, error) {
	return nil, nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRunGenerators_SourceAll(t *testing.T) {
	gen1 := &mockGenerator{name: "gen-1", slides: []domain.Slide{{ExternalID: "s1"}}}
	gen2 := &mockGenerator{name: "gen-2", slides: []domain.Slide{{ExternalID: "s2"}}}
	bc := &mockBroadcaster{}
	repo := &mockSlideRepo{
		result: &repository.SlideSyncResult{
			Created: []domain.Slide{{ExternalID: "created"}},
		},
	}
	engine := &Engine{
		generators:     []SlideGenerator{gen1, gen2},
		slideRepo:      repo,
		sseBroadcaster: bc,
		logger:         newTestLogger(),
	}

	engine.runGenerators(context.Background(), "all")

	// Both generators called Sync, each Sync returns 1 Created → 2 CREATE events
	assert.Equal(t, 2, len(bc.events), "both generators should be called")
	assert.Equal(t, domain.EventCreate, bc.events[0])
	assert.Equal(t, domain.EventCreate, bc.events[1])
}

func TestRunGenerators_SpecificSource(t *testing.T) {
	gen1 := &mockGenerator{name: "gen-1", slides: []domain.Slide{{ExternalID: "s1"}}}
	gen2 := &mockGenerator{name: "gen-2", slides: []domain.Slide{{ExternalID: "s2"}}}
	bc := &mockBroadcaster{}
	repo := &mockSlideRepo{
		result: &repository.SlideSyncResult{
			Created: []domain.Slide{{ExternalID: "s1"}},
		},
	}
	engine := &Engine{
		generators:     []SlideGenerator{gen1, gen2},
		slideRepo:      repo,
		sseBroadcaster: bc,
		logger:         newTestLogger(),
	}

	engine.runGenerators(context.Background(), "gen-1")

	assert.Equal(t, 1, len(bc.events), "only gen-1 should be called")
}

func TestRunGenerators_GeneratorReturnsError(t *testing.T) {
	gen1 := &mockGenerator{name: "gen-1", err: assert.AnError}
	gen2 := &mockGenerator{name: "gen-2", slides: []domain.Slide{{ExternalID: "s2"}}}
	bc := &mockBroadcaster{}
	repo := &mockSlideRepo{
		result: &repository.SlideSyncResult{
			Created: []domain.Slide{{ExternalID: "s2"}},
		},
	}
	engine := &Engine{
		generators:     []SlideGenerator{gen1, gen2},
		slideRepo:      repo,
		sseBroadcaster: bc,
		logger:         newTestLogger(),
	}

	engine.runGenerators(context.Background(), "all")

	// gen-1 errored, gen-2 should still run
	assert.Equal(t, 1, len(bc.events), "gen-2 should still be called after gen-1 error")
	assert.Equal(t, domain.EventCreate, bc.events[0])
}

func TestRunGenerators_SyncReturnsError(t *testing.T) {
	gen := &mockGenerator{name: "gen-1", slides: []domain.Slide{{ExternalID: "s1"}}}
	bc := &mockBroadcaster{}
	repo := &mockSlideRepo{err: assert.AnError}
	engine := &Engine{
		generators:     []SlideGenerator{gen},
		slideRepo:      repo,
		sseBroadcaster: bc,
		logger:         newTestLogger(),
	}

	// Should not panic; error is logged
	engine.runGenerators(context.Background(), "all")

	assert.Empty(t, bc.events, "no broadcast on sync error")
}

func TestRunGenerators_BroadcastCreateUpdateDelete(t *testing.T) {
	gen := &mockGenerator{name: "gen-1", slides: []domain.Slide{{ExternalID: "s1"}}}
	bc := &mockBroadcaster{}
	repo := &mockSlideRepo{
		result: &repository.SlideSyncResult{
			Created: []domain.Slide{{ExternalID: "created"}},
			Updated: []domain.Slide{{ExternalID: "updated"}},
			Deleted: []domain.Slide{{ID: 42}},
		},
	}
	engine := &Engine{
		generators:     []SlideGenerator{gen},
		slideRepo:      repo,
		sseBroadcaster: bc,
		logger:         newTestLogger(),
	}

	engine.runGenerators(context.Background(), "all")

	assert.Len(t, bc.events, 3)
	assert.Equal(t, domain.EventCreate, bc.events[0])
	assert.Equal(t, domain.EventUpdate, bc.events[1])
	assert.Equal(t, domain.EventDelete, bc.events[2])
}

func TestTrigger_Success(t *testing.T) {
	engine := &Engine{
		triggerCh: make(chan string, 1),
		logger:    newTestLogger(),
	}

	engine.Trigger("test-source")

	select {
	case src := <-engine.triggerCh:
		assert.Equal(t, "test-source", src)
	default:
		t.Fatal("expected trigger signal on channel")
	}
}

func TestTrigger_ChannelFull(t *testing.T) {
	engine := &Engine{
		triggerCh: make(chan string, 1),
		logger:    newTestLogger(),
	}

	// Fill the channel
	engine.triggerCh <- "first"

	// This should not block
	engine.Trigger("second")

	// The channel still has "first" (second was dropped)
	select {
	case src := <-engine.triggerCh:
		assert.Equal(t, "first", src)
	default:
		t.Fatal("expected first trigger on channel")
	}
}

func TestNewEngine(t *testing.T) {
	repo := &mockSlideRepo{}
	bc := &mockBroadcaster{}
	logger := newTestLogger()
	gen := &mockGenerator{name: "test-gen"}

	engine := NewEngine(repo, bc, logger, gen)

	require.NotNil(t, engine)
	assert.Equal(t, 1*time.Minute, engine.interval)
	assert.Len(t, engine.generators, 1)
	assert.NotNil(t, engine.triggerCh)
}
