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

type NewsSyncResult struct {
	Created []domain.News
	Updated []domain.News
	Deleted []domain.News
}

type NewsRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.News, error)
	List(ctx context.Context, params NewsListParams, filters NewsListFilters) ([]domain.News, int64, error)
	Save(ctx context.Context, news *domain.News) error
	Delete(ctx context.Context, source, externalID string) error

	GetAll(ctx context.Context) ([]domain.News, error)

	Sync(ctx context.Context, source string, newsItems []domain.News) (*NewsSyncResult, error)
}
