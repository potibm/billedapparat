package domain

import (
	"sort"
	"time"
)

type TimetableEvent struct {
	ID          int64     `json:"id"`
	ExternalID  string    `json:"external_id"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ExternalURL string    `json:"external_url"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsHidden    bool      `json:"hidden"`
	Category    *Category `json:"category,omitempty"`
	Location    *Location `json:"location,omitempty"`
}

type Timetable []TimetableEvent

type TimetableDay struct {
	DateStr string
	Date    time.Time
	Events  Timetable
}

type Category struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Location struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (t Timetable) Sort() {
	sort.Slice(t, func(i, j int) bool {
		if t[i].StartTime.Equal(t[j].StartTime) {
			return t[i].EndTime.Before(t[j].EndTime)
		}

		return t[i].StartTime.Before(t[j].StartTime)
	})
}

func (t Timetable) GroupByDate() []TimetableDay {
	if len(t) == 0 {
		return nil
	}

	sorted := append(Timetable(nil), t...)
	sorted.Sort()

	var (
		days       []TimetableDay
		currentDay *TimetableDay
	)

	for _, event := range sorted {
		dateStr := event.StartTime.Format(time.DateOnly)

		if currentDay == nil || currentDay.DateStr != dateStr {
			y, m, d := event.StartTime.Date()
			days = append(days, TimetableDay{
				DateStr: dateStr,
				Date:    time.Date(y, m, d, 0, 0, 0, 0, event.StartTime.Location()),
				Events:  Timetable{},
			})
			currentDay = &days[len(days)-1]
		}

		currentDay.Events = append(currentDay.Events, event)
	}

	return days
}

func (t Timetable) Chunk(size int) []Timetable {
	if size <= 0 {
		return []Timetable{t}
	}

	var chunks []Timetable

	for i := 0; i < len(t); i += size {
		end := i + size
		if end > len(t) {
			end = len(t)
		}

		chunks = append(chunks, t[i:end])
	}

	return chunks
}

// ChunkEvenly splits t into as few chunks as possible without any chunk
// exceeding max items. Chunks are as evenly sized as possible (differ by
// at most one item; the first chunks get the extra item to preserve
// chronological order).
//
// Precondition: chunkSize > 0 (guaranteed by the config layer in production).
// If chunkSize <= 0, the timetable is returned as a single chunk.
func (t Timetable) ChunkEvenly(chunkSize int) []Timetable {
	n := len(t)

	if n == 0 {
		return nil
	}

	if chunkSize <= 0 || chunkSize >= n {
		return []Timetable{t}
	}

	chunks := (n + chunkSize - 1) / chunkSize // ceil(n / chunkSize) — minimum chunks needed
	base := n / chunks
	remainder := n % chunks // first `remainder` chunks get one extra item

	result := make([]Timetable, 0, chunks)
	start := 0

	for i := 0; i < chunks; i++ {
		size := base
		if i < remainder {
			size++
		}

		result = append(result, t[start:start+size])
		start += size
	}

	return result
}
