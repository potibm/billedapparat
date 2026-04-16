package repository

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type SlideRepository interface {
	Save(ctx context.Context, slide *domain.Slide) error
	GetActive(ctx context.Context) ([]domain.Slide, error)
	Delete(ctx context.Context, id int64) error
	AdminList(ctx context.Context, params AdminListParams) ([]domain.Slide, int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Slide, error)
	GetAllMediaURLs(ctx context.Context) ([]string, error)
}

type AdminListParams struct {
	Offset int
	Limit  int
	Sort   string
	Order  string
	Type   domain.SlideType
}
