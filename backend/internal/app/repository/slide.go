package repository

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type AdminListFilters struct {
	Query    *string
	Status   *string
	Priority *int32
	ID       *int64
}

type SlideRepository interface {
	Save(ctx context.Context, slide *domain.Slide) error
	GetActive(ctx context.Context) ([]domain.Slide, error)
	Delete(ctx context.Context, id int64) error
	AdminList(ctx context.Context, params AdminListParams, filters AdminListFilters) ([]domain.Slide, int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Slide, error)
	GetAllMediaURLs(ctx context.Context) ([]string, error)
	SlideExists(source, externalID string, subID *int) (bool, error)
	FindLocalURLByOriginalURL(ctx context.Context, originalURL string) (string, bool)
}

type AdminListParams struct {
	Offset int
	Limit  int
	Sort   string
	Order  string
	Type   domain.SlideType
}
