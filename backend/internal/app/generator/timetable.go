package generator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
)

const defaultEntriesPerSlide = 5

type timetableGenerator struct {
	timetableRepo   repository.TimetableEventRepository
	logger          *slog.Logger
	entriesPerSlide int
}

func NewTimetableGenerator(timetableRepo repository.TimetableEventRepository, logger *slog.Logger) *timetableGenerator {
	return &timetableGenerator{
		timetableRepo:   timetableRepo,
		logger:          logger,
		entriesPerSlide: defaultEntriesPerSlide,
	}
}

func (g *timetableGenerator) Name() string {
	return "timetable-generator"
}

func (g *timetableGenerator) Generate(ctx context.Context) ([]domain.Slide, error) {
	events, err := g.timetableRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	dailyGroups := domain.Timetable(events).GroupByDate()

	var slides []domain.Slide

	for _, dayGroup := range dailyGroups {
		chunks := dayGroup.Events.Chunk(g.entriesPerSlide)

		for pageIndex, chunk := range chunks {
			slides = append(slides, g.timetableToSlide(chunk, dayGroup.Date, dayGroup.DateStr, pageIndex, len(chunks)))
		}
	}

	g.logger.Info("Generated slides from timetable", "count", len(slides))

	return slides, nil
}

func (g *timetableGenerator) timetableToSlide(
	t domain.Timetable,
	date time.Time,
	dateStr string,
	subID, totalPages int,
) domain.Slide {
	subIDRef := subID

	title := "Timetable " + date.Format("Monday")
	if totalPages > 1 {
		title += fmt.Sprintf(" %d/%d", subID+1, totalPages)
	}

	body := "# " + title

	body += "\n\n"
	body += "| Start Time | End Time | Category | Title | Location |\n"
	body += "| --- | --- | --- | --- | --- |\n"

	for _, event := range t {
		location := ""
		if event.Location != nil {
			location = event.Location.Name
		}

		endTimeStr := ""
		if event.StartTime != event.EndTime {
			endTimeStr = event.EndTime.Format("15:04")
		}

		category := ""
		if event.Category != nil {
			category = event.Category.Name
		}

		body += fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			event.StartTime.Format("15:04"),
			endTimeStr,
			category,
			event.Title,
			location,
		)
	}

	slide := domain.Slide{
		Source:        g.Name(),
		ExternalID:    dateStr,
		ExternalSubID: &subIDRef,
		Status:        domain.StatusActive,
		Content: domain.Content{
			Title: title,
			Body:  body,
			Type:  domain.TypeTimetable,
		},
	}

	return slide
}
