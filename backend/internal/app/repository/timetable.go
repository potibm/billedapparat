package repository

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type TimetableEventListParams struct {
	ListParams
}

type TimetableEventListFilters struct {
	Query      *string
	Source     *string
	IsHidden   *bool
	ExternalID *string
}

type TimetableEventSyncResult struct {
	Created []domain.TimetableEvent
	Updated []domain.TimetableEvent
	Deleted []domain.TimetableEvent
}

type TimetableEventRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.TimetableEvent, error)
	List(
		ctx context.Context,
		params TimetableEventListParams,
		filters TimetableEventListFilters,
	) (domain.Timetable, int64, error)
	Save(ctx context.Context, event *domain.TimetableEvent) error
	Delete(ctx context.Context, source, externalID string) error

	GetAll(ctx context.Context) (domain.Timetable, error)
	GetActive(ctx context.Context) (domain.Timetable, error)

	Sync(ctx context.Context, source string, timetableItems []domain.TimetableEvent) (*TimetableEventSyncResult, error)
}
