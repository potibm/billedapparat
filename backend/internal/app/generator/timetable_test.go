package generator

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTime(hour, minute int) time.Time {
	return time.Date(2026, time.May, 25, hour, minute, 0, 0, time.UTC)
}

func TestTimetableToSlide_SinglePage(t *testing.T) {
	g := &timetableGenerator{logger: slog.Default()}

	date := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	events := domain.Timetable{
		{
			ExternalID: "evt-1",
			Title:      "Morning Yoga",
			StartTime:  makeTime(9, 0),
			EndTime:    makeTime(10, 0),
			Location:   &domain.Location{Name: "Studio 1"},
			Category:   &domain.Category{Name: "Sports"},
		},
	}

	slide := g.timetableToSlide(events, date, "2026-05-25", 0, 1)

	assert.Equal(t, "Monday", slide.Content.Title)
	assert.NotContains(t, slide.Content.Title, "/")
	assert.Equal(t, domain.TypeTimetable, slide.Content.Type)
	assert.Equal(t, domain.StatusActive, slide.Status)
	assert.Equal(t, "2026-05-25", slide.ExternalID)
	assert.Equal(t, 0, *slide.ExternalSubID)
	assert.Contains(t, slide.Content.Body, "| Morning Yoga |")
	assert.Contains(t, slide.Content.Body, "| 09:00 | 10:00 | Sports |  | Morning Yoga | Studio 1 |")
}

func TestTimetableToSlide_MultiPage(t *testing.T) {
	g := &timetableGenerator{logger: slog.Default()}

	date := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	events := domain.Timetable{
		{
			ExternalID: "evt-1",
			Title:      "Event",
			StartTime:  makeTime(14, 0),
			EndTime:    makeTime(15, 0),
		},
	}

	slide := g.timetableToSlide(events, date, "2026-05-25", 1, 2)

	// subID 1 + 1 = page 2, total 2 → "2/2"
	assert.Equal(t, "Monday 2/2", slide.Content.Title)
	assert.Equal(t, 1, *slide.ExternalSubID)
}

func TestTimetableToSlide_NilLocation(t *testing.T) {
	g := &timetableGenerator{logger: slog.Default()}

	date := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	events := domain.Timetable{
		{
			ExternalID: "evt-1",
			Title:      "Event Without Location",
			StartTime:  makeTime(9, 0),
			EndTime:    makeTime(10, 0),
		},
	}

	slide := g.timetableToSlide(events, date, "2026-05-25", 0, 1)

	assert.Contains(t, slide.Content.Body, "Event Without Location")
	// The location column should be empty at the end of the row
	assert.Contains(t, slide.Content.Body, "| Event Without Location |  |")
}

func TestTimetableToSlide_NilCategory(t *testing.T) {
	g := &timetableGenerator{logger: slog.Default()}

	date := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	events := domain.Timetable{
		{
			ExternalID: "evt-1",
			Title:      "Event Without Category",
			StartTime:  makeTime(9, 0),
			EndTime:    makeTime(10, 0),
			Location:   &domain.Location{Name: "Room A"},
		},
	}

	slide := g.timetableToSlide(events, date, "2026-05-25", 0, 1)

	assert.Contains(t, slide.Content.Body, "| Event Without Category |")
	assert.Contains(t, slide.Content.Body, "Room A")
}

func TestTimetableToSlide_StartTimeEqualsEndTime(t *testing.T) {
	g := &timetableGenerator{logger: slog.Default()}

	date := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	events := domain.Timetable{
		{
			ExternalID: "evt-1",
			Title:      "All Day Event",
			StartTime:  makeTime(9, 0),
			EndTime:    makeTime(9, 0),
		},
	}

	slide := g.timetableToSlide(events, date, "2026-05-25", 0, 1)

	// End time should be empty when start == end
	assert.NotContains(t, slide.Content.Body, "| 09:00 | 09:00 |", "end time should be empty")
	assert.Contains(t, slide.Content.Body, "| 09:00 |  | ")
}

func TestTimetableToSlide_AllFields(t *testing.T) {
	g := &timetableGenerator{logger: slog.Default()}

	date := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	events := domain.Timetable{
		{
			ExternalID: "evt-1",
			Title:      "Full Event",
			StartTime:  makeTime(10, 0),
			EndTime:    makeTime(12, 0),
			Location:   &domain.Location{Name: "Main Hall", Address: "123 Main St"},
			Category:   &domain.Category{Name: "Workshop", Color: "#ff0000"},
		},
	}

	slide := g.timetableToSlide(events, date, "2026-05-25", 0, 1)

	assert.Equal(t, domain.TypeTimetable, slide.Content.Type)
	assert.Equal(t, domain.StatusActive, slide.Status)

	expectedRow := "| 10:00 | 12:00 | Workshop | #ff0000 | Full Event | Main Hall |"
	assert.Contains(t, slide.Content.Body, expectedRow)
}

// mockTimetableRepo implements repository.TimetableEventRepository for testing.
type mockTimetableRepo struct {
	events domain.Timetable
	err    error
}

func (m *mockTimetableRepo) GetActive(ctx context.Context) (domain.Timetable, error) {
	return m.events, m.err
}

func (m *mockTimetableRepo) GetAll(ctx context.Context) (domain.Timetable, error) {
	return m.events, m.err
}

func (m *mockTimetableRepo) Save(ctx context.Context, event *domain.TimetableEvent) error {
	return m.err
}

func (m *mockTimetableRepo) Delete(ctx context.Context, source, externalID string) error {
	return m.err
}

func (m *mockTimetableRepo) GetByID(ctx context.Context, id int64) (*domain.TimetableEvent, error) {
	return nil, m.err
}

func (m *mockTimetableRepo) List(
	ctx context.Context,
	params repository.TimetableEventListParams,
	filters repository.TimetableEventListFilters,
) (domain.Timetable, int64, error) {
	return m.events, int64(len(m.events)), m.err
}

func (m *mockTimetableRepo) Sync(
	ctx context.Context,
	source string,
	items []domain.TimetableEvent,
) (*repository.TimetableEventSyncResult, error) {
	return nil, m.err
}

func TestGenerate_EmptyEvents(t *testing.T) {
	repo := &mockTimetableRepo{events: domain.Timetable{}}
	g := NewTimetableGenerator(repo, slog.Default(), 7)

	slides, err := g.Generate(context.Background())

	require.NoError(t, err)
	assert.Empty(t, slides)
}

func TestGenerate_OneDayWithinLimit(t *testing.T) {
	events := domain.Timetable{
		{ExternalID: "e1", Title: "Event 1", StartTime: makeTime(9, 0), EndTime: makeTime(10, 0)},
		{ExternalID: "e2", Title: "Event 2", StartTime: makeTime(11, 0), EndTime: makeTime(12, 0)},
	}
	repo := &mockTimetableRepo{events: events}
	// 5 entries per slide → both events fit on one slide
	g := NewTimetableGenerator(repo, slog.Default(), 5)

	slides, err := g.Generate(context.Background())

	require.NoError(t, err)
	assert.Len(t, slides, 1)
}

func TestGenerate_OneDayExceedingLimit(t *testing.T) {
	events := domain.Timetable{
		{ExternalID: "e1", Title: "E1", StartTime: makeTime(9, 0), EndTime: makeTime(10, 0)},
		{ExternalID: "e2", Title: "E2", StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
		{ExternalID: "e3", Title: "E3", StartTime: makeTime(11, 0), EndTime: makeTime(12, 0)},
		{ExternalID: "e4", Title: "E4", StartTime: makeTime(12, 0), EndTime: makeTime(13, 0)},
	}
	repo := &mockTimetableRepo{events: events}
	// 2 entries per slide → 4 events → 2 slides
	g := NewTimetableGenerator(repo, slog.Default(), 2)

	slides, err := g.Generate(context.Background())

	require.NoError(t, err)
	assert.Len(t, slides, 2)
	// first slide should have pagination
	assert.Contains(t, slides[0].Content.Title, "1/2")
	assert.Contains(t, slides[1].Content.Title, "2/2")
}

func TestGenerate_MultipleDays(t *testing.T) {
	events := domain.Timetable{
		{
			ExternalID: "e1", Title: "Day1 Event",
			StartTime: makeTime(9, 0),
			EndTime:   makeTime(10, 0),
		},
		{
			ExternalID: "e2", Title: "Day2 Event",
			StartTime: time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, time.May, 26, 11, 0, 0, 0, time.UTC),
		},
	}
	repo := &mockTimetableRepo{events: events}
	g := NewTimetableGenerator(repo, slog.Default(), 7)

	slides, err := g.Generate(context.Background())

	require.NoError(t, err)
	// group by date → 2 days → 2 slides
	assert.Len(t, slides, 2)
}

func TestGenerate_RepoError(t *testing.T) {
	repo := &mockTimetableRepo{
		err: assert.AnError,
	}
	g := NewTimetableGenerator(repo, slog.Default(), 7)

	slides, err := g.Generate(context.Background())

	assert.Error(t, err)
	assert.Nil(t, slides)
}
