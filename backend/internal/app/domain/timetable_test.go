package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func makeTime(hour, minute int) time.Time {
	return time.Date(2026, time.May, 24, hour, minute, 0, 0, time.UTC)
}

func makeTimeOnDate(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func TestTimetable_Sort(t *testing.T) {
	tests := []struct {
		name     string
		input    Timetable
		expected Timetable
	}{
		{
			name:     "empty timetable",
			input:    Timetable{},
			expected: Timetable{},
		},
		{
			name: "already sorted",
			input: Timetable{
				{ID: 1, StartTime: makeTime(9, 0), EndTime: makeTime(10, 0)},
				{ID: 2, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
			},
			expected: Timetable{
				{ID: 1, StartTime: makeTime(9, 0), EndTime: makeTime(10, 0)},
				{ID: 2, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
			},
		},
		{
			name: "reverse order",
			input: Timetable{
				{ID: 2, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
				{ID: 1, StartTime: makeTime(9, 0), EndTime: makeTime(10, 0)},
			},
			expected: Timetable{
				{ID: 1, StartTime: makeTime(9, 0), EndTime: makeTime(10, 0)},
				{ID: 2, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
			},
		},
		{
			name: "same start time, sort by end time",
			input: Timetable{
				{ID: 2, StartTime: makeTime(10, 0), EndTime: makeTime(12, 0)},
				{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
			},
			expected: Timetable{
				{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
				{ID: 2, StartTime: makeTime(10, 0), EndTime: makeTime(12, 0)},
			},
		},
		{
			name: "mixed order",
			input: Timetable{
				{ID: 3, StartTime: makeTime(14, 0), EndTime: makeTime(15, 0)},
				{ID: 1, StartTime: makeTime(9, 0), EndTime: makeTime(10, 0)},
				{ID: 2, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
			},
			expected: Timetable{
				{ID: 1, StartTime: makeTime(9, 0), EndTime: makeTime(10, 0)},
				{ID: 2, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
				{ID: 3, StartTime: makeTime(14, 0), EndTime: makeTime(15, 0)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Sort()
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestTimetable_GroupByDate(t *testing.T) {
	tests := []struct {
		name     string
		input    Timetable
		expected []TimetableDay
	}{
		{
			name:     "empty timetable returns nil",
			input:    Timetable{},
			expected: nil,
		},
		{
			name: "single event",
			input: Timetable{
				{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
			},
			expected: []TimetableDay{
				{
					DateStr: "2026-05-24",
					Date:    time.Date(2026, time.May, 24, 0, 0, 0, 0, time.UTC),
					Events: Timetable{
						{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
					},
				},
			},
		},
		{
			name: "multiple events on same day",
			input: Timetable{
				{ID: 2, StartTime: makeTime(14, 0), EndTime: makeTime(15, 0)},
				{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
			},
			expected: []TimetableDay{
				{
					DateStr: "2026-05-24",
					Date:    time.Date(2026, time.May, 24, 0, 0, 0, 0, time.UTC),
					Events: Timetable{
						{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
						{ID: 2, StartTime: makeTime(14, 0), EndTime: makeTime(15, 0)},
					},
				},
			},
		},
		{
			name: "events on different days",
			input: Timetable{
				{
					ID: 2, StartTime: makeTimeOnDate(2026, time.May, 25, 10, 0),
					EndTime: makeTimeOnDate(2026, time.May, 25, 11, 0),
				},
				{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
			},
			expected: []TimetableDay{
				{
					DateStr: "2026-05-24",
					Date:    time.Date(2026, time.May, 24, 0, 0, 0, 0, time.UTC),
					Events: Timetable{
						{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
					},
				},
				{
					DateStr: "2026-05-25",
					Date:    time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC),
					Events: Timetable{
						{
							ID: 2, StartTime: makeTimeOnDate(2026, time.May, 25, 10, 0),
							EndTime: makeTimeOnDate(2026, time.May, 25, 11, 0),
						},
					},
				},
			},
		},
		{
			name: "unsorted input gets sorted",
			input: Timetable{
				{ID: 3, StartTime: makeTime(16, 0), EndTime: makeTime(17, 0)},
				{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
				{ID: 2, StartTime: makeTime(14, 0), EndTime: makeTime(15, 0)},
			},
			expected: []TimetableDay{
				{
					DateStr: "2026-05-24",
					Date:    time.Date(2026, time.May, 24, 0, 0, 0, 0, time.UTC),
					Events: Timetable{
						{ID: 1, StartTime: makeTime(10, 0), EndTime: makeTime(11, 0)},
						{ID: 2, StartTime: makeTime(14, 0), EndTime: makeTime(15, 0)},
						{ID: 3, StartTime: makeTime(16, 0), EndTime: makeTime(17, 0)},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.GroupByDate()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestTimetable_Chunk(t *testing.T) {
	tests := []struct {
		name     string
		input    Timetable
		size     int
		expected []Timetable
	}{
		{
			name:     "empty timetable",
			input:    Timetable{},
			size:     2,
			expected: nil,
		},
		{
			name: "size zero returns whole timetable",
			input: Timetable{
				{ID: 1},
				{ID: 2},
			},
			size:     0,
			expected: []Timetable{{{ID: 1}, {ID: 2}}},
		},
		{
			name: "negative size returns whole timetable",
			input: Timetable{
				{ID: 1},
				{ID: 2},
			},
			size:     -1,
			expected: []Timetable{{{ID: 1}, {ID: 2}}},
		},
		{
			name: "size equal to length",
			input: Timetable{
				{ID: 1},
				{ID: 2},
			},
			size:     2,
			expected: []Timetable{{{ID: 1}, {ID: 2}}},
		},
		{
			name: "size larger than length",
			input: Timetable{
				{ID: 1},
				{ID: 2},
			},
			size:     5,
			expected: []Timetable{{{ID: 1}, {ID: 2}}},
		},
		{
			name: "size one",
			input: Timetable{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			size:     1,
			expected: []Timetable{{{ID: 1}}, {{ID: 2}}, {{ID: 3}}},
		},
		{
			name: "even split",
			input: Timetable{
				{ID: 1},
				{ID: 2},
				{ID: 3},
				{ID: 4},
			},
			size:     2,
			expected: []Timetable{{{ID: 1}, {ID: 2}}, {{ID: 3}, {ID: 4}}},
		},
		{
			name: "uneven split",
			input: Timetable{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			size:     2,
			expected: []Timetable{{{ID: 1}, {ID: 2}}, {{ID: 3}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Chunk(tt.size)
			assert.Equal(t, tt.expected, got)
		})
	}
}
