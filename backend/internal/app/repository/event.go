package repository

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type EventListParams struct {
	ListParams
}

type EventListFilters struct {
	Query    *string
	Source   *string
	IsHidden *bool
}

type TimetableEventRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.TimetableEvent, error)
	List(ctx context.Context, params EventListParams, filters EventListFilters) ([]domain.TimetableEvent, int64, error)
	Save(ctx context.Context, event *domain.TimetableEvent) error
	Delete(ctx context.Context, id int64) error

	GetAll(ctx context.Context) ([]domain.TimetableEvent, error)
}
