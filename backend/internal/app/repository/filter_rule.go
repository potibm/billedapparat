package repository

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type FilterRuleRepository interface {
	List(ctx context.Context, limit, offset int) ([]domain.FilterRule, int64, error)
	GetByID(ctx context.Context, id uint) (*domain.FilterRule, error)
	Create(ctx context.Context, rule *domain.FilterRule) error
	Update(ctx context.Context, rule *domain.FilterRule) error
	Delete(ctx context.Context, id uint) error
}
