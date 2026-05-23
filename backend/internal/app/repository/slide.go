package repository

import (
	"context"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type ListParams struct {
	Offset int
	Limit  int
	Sort   string
	Order  string
}

type SlideListFilters struct {
	Query    *string
	Status   *string
	Priority *int32
	ID       *int64
	Source   *string
}

type SlideListParams struct {
	ListParams

	Type domain.SlideType
}

type SlideSyncResult struct {
	Created []domain.Slide
	Updated []domain.Slide
	Deleted []domain.Slide
}

//nolint:interfacebloat // the repository interface is a central part of the application
type SlideRepository interface {
	Save(ctx context.Context, slide *domain.Slide) error
	GetActive(ctx context.Context) ([]domain.Slide, error)
	Delete(ctx context.Context, id int64) error
	AdminList(ctx context.Context, params SlideListParams, filters SlideListFilters) ([]domain.Slide, int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Slide, error)
	MarkAsDeleted(ctx context.Context, source, externalID string) error
	SlideExists(source, externalID string, subID *int) (bool, error)
	GetAllMediaURLs(ctx context.Context) ([]string, error)
	FindLocalURLByOriginalURL(ctx context.Context, originalURL string) (string, bool)
	FindExpiredSlidesByType(ctx context.Context, slideType string, cutoff time.Time) ([]domain.Slide, error)
	Sync(ctx context.Context, source string, newSlides []domain.Slide) (*SlideSyncResult, error)
}
