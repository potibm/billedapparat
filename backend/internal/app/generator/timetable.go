package generator

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
)

type timetableGenerator struct {
	timetableRepo      repository.TimetableEventRepository
	logger             *slog.Logger
	maxEntriesPerSlide int
}

func NewTimetableGenerator(
	timetableRepo repository.TimetableEventRepository,
	logger *slog.Logger,
	maxEntriesPerSlide int,
) *timetableGenerator {
	return &timetableGenerator{
		timetableRepo:      timetableRepo,
		logger:             logger,
		maxEntriesPerSlide: maxEntriesPerSlide,
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
		chunks := dayGroup.Events.ChunkEvenly(g.maxEntriesPerSlide)

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

	title := date.Format("Monday")
	if totalPages > 1 {
		title += fmt.Sprintf(" %d/%d", subID+1, totalPages)
	}

	body := "| Start Time | End Time | Category | Category Color | Title | Location |\n"
	body += "| --- | --- | --- | --- | --- | --- |\n"

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
		categoryColor := ""

		if event.Category != nil {
			category = event.Category.Name
			categoryColor = event.Category.Color
		}

		body += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			escapePipes(event.StartTime.Format("15:04")),
			escapePipes(endTimeStr),
			escapePipes(category),
			escapePipes(categoryColor),
			escapePipes(event.Title),
			escapePipes(location),
		)
	}

	originCreatedAt := date
	if len(t) > 0 {
		originCreatedAt = t[0].StartTime
	}

	slide := domain.Slide{
		Source:          g.Name(),
		ExternalID:      dateStr,
		ExternalSubID:   &subIDRef,
		Status:          domain.StatusActive,
		OriginCreatedAt: originCreatedAt,
		Content: domain.Content{
			Title: title,
			Body:  body,
			Type:  domain.TypeTimetable,
		},
	}

	return slide
}

func escapePipes(s string) string {
	return strings.ReplaceAll(s, "|", "&#124;")
}
