package gorm

import (
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type dbTimetableEvent struct {
	GormModel

	Source     string `gorm:"uniqueIndex:tte_idx_ext"`
	ExternalID string `gorm:"uniqueIndex:tte_idx_ext"`

	Title       string
	Description string
	ExternalURL string
	StartTime   time.Time
	EndTime     time.Time
	IsHidden    bool

	CategoryName  string
	CategoryColor string

	LocationName    string
	LocationAddress string
}

func (dbTimetableEvent) TableName() string {
	return "timetable_events"
}

func fromDomainTimetableEvent(e *domain.TimetableEvent) *dbTimetableEvent {
	db := &dbTimetableEvent{
		GormModel: GormModel{ID: e.ID},

		Source:     e.Source,
		ExternalID: e.ExternalID,

		Title:       e.Title,
		Description: e.Description,
		ExternalURL: e.ExternalURL,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
		IsHidden:    e.IsHidden,
	}

	if e.Category != nil {
		db.CategoryName = e.Category.Name
		db.CategoryColor = e.Category.Color
	}

	if e.Location != nil {
		db.LocationName = e.Location.Name
		db.LocationAddress = e.Location.Address
	}

	return db
}

func (n *dbTimetableEvent) toDomain() *domain.TimetableEvent {
	result := &domain.TimetableEvent{
		ID: n.ID,

		Source:     n.Source,
		ExternalID: n.ExternalID,

		Title:       n.Title,
		Description: n.Description,
		ExternalURL: n.ExternalURL,
		StartTime:   n.StartTime,
		EndTime:     n.EndTime,
		IsHidden:    n.IsHidden,
	}

	if n.CategoryName != "" || n.CategoryColor != "" {
		result.Category = &domain.Category{
			Name:  n.CategoryName,
			Color: n.CategoryColor,
		}
	}

	if n.LocationName != "" || n.LocationAddress != "" {
		result.Location = &domain.Location{
			Name:    n.LocationName,
			Address: n.LocationAddress,
		}
	}

	return result
}

func toDomainTimetableEventList(events []dbTimetableEvent) domain.Timetable {
	result := make(domain.Timetable, len(events))
	for i, e := range events {
		result[i] = *e.toDomain()
	}

	return result
}
