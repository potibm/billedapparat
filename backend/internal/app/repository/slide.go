package repository

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type SlideRepository interface {
	Save(ctx context.Context, slide *domain.Slide) error
	GetActive(ctx context.Context) ([]domain.Slide, error)
	Delete(ctx context.Context, id uint) error
}
