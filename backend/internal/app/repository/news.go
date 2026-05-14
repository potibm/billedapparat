package repository

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type NewsListParams struct {
	ListParams
}

type NewsListFilters struct {
	Query    *string
	IsUrgent *bool
	IsHidden *bool
}

type NewsRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.News, error)
	List(ctx context.Context, params NewsListParams, filters NewsListFilters) ([]domain.News, int64, error)
	Save(ctx context.Context, news *domain.News) error
	Delete(ctx context.Context, id int64) error

	GetAll(ctx context.Context) ([]domain.News, error)
}
