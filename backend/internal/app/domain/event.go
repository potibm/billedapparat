package domain

import "time"

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

type Category struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Location struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
